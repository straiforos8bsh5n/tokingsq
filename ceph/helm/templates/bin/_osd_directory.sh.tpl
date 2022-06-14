#!/bin/bash
set -ex

function is_integer {
  # This function is about saying if the passed argument is an integer
  # Supports also negative integers
  # We use $@ here to consider everything given as parameter and not only the
  # first one : that's mainly for splited strings like "10 10"
  [[ $@ =~ ^-?[0-9]+$ ]]
}

function osd_directory {
  local test_luminous=$(ceph -v | egrep -q "12.2|luminous"; echo $?)
  if [[ ${test_luminous} -ne 0 ]]; then
      log "ERROR- need Luminous release"
      exit 1
  fi

  if [[ ! -d /var/lib/ceph/osd ]]; then
    log "ERROR- could not find the osd directory, did you bind mount the OSD data directory?"
    log "ERROR- use -v <host_osd_data_dir>:/var/lib/ceph/osd"
    exit 1
  fi

  if [ -z "${HOSTNAME}" ]; then
    log "HOSTNAME not set; This will prevent to add an OSD into the CRUSH map"
    exit 1
  fi

  chown ceph. /var/log/ceph

  # check if anything is present, if not, create an osd and its directory
  if [[ -n "$(find /var/lib/ceph/osd -prune -empty)" ]]; then
    log "Creating osd"
    UUID=$(uuidgen)
    OSD_SECRET=$(ceph-authtool --gen-print-key)
    OSD_ID=$(echo "{\"cephx_secret\": \"${OSD_SECRET}\"}" | ceph osd new ${UUID} -i - -n client.bootstrap-osd -k "$OSD_BOOTSTRAP_KEYRING")
    if is_integer "$OSD_ID"; then
      log "OSD created with ID: ${OSD_ID}"
    else
      log "OSD creation failed: ${OSD_ID}"
      exit 1
    fi

    OSD_PATH=$(get_osd_path "$OSD_ID")
    if [ -n "${JOURNAL_DIR}" ]; then
       OSD_J="${JOURNAL_DIR}/journal.${OSD_ID}"
       chown -R ceph. ${JOURNAL_DIR}
    else
       if [ -n "${JOURNAL}" ]; then
          OSD_J=${JOURNAL}
          chown -R ceph. $(dirname ${JOURNAL_DIR})
       else
          OSD_J=${OSD_PATH}/journal
       fi
    fi
    # create the folder and own it
    mkdir -p "$OSD_PATH"
    chown "${CHOWN_OPT[@]}" ceph. "$OSD_PATH"
    log "created folder $OSD_PATH"
    # write the secret to the osd keyring file
    ceph-authtool --create-keyring ${OSD_PATH}/keyring --name osd.${OSD_ID} --add-key ${OSD_SECRET}
    OSD_KEYRING="$OSD_PATH/keyring"
    # init data directory
    ceph-osd -i ${OSD_ID} --mkfs --osd-uuid ${UUID} --mkjournal --osd-journal ${OSD_J} --setuser ceph --setgroup ceph
    # add the osd to the crush map
    OSD_WEIGHT=$(df -P -k $OSD_PATH | tail -1 | awk '{ d= $2/1073741824 ; r = sprintf("%.2f", d); print r }')
    ceph --name=osd.${OSD_ID} --keyring=${OSD_KEYRING} osd crush create-or-move -- ${OSD_ID} ${OSD_WEIGHT} ${CRUSH_LOCATION}
  fi

  # create the directory and an empty Procfile
  mkdir -p /etc/forego/"${CLUSTER}"
  echo "" > /etc/forego/"${CLUSTER}"/Procfile

  for OSD_ID in $(ls /var/lib/ceph/osd | sed 's/.*-//'); do
    OSD_PATH=$(get_osd_path "$OSD_ID")
    OSD_KEYRING="$OSD_PATH/keyring"
    if [ -n "${JOURNAL_DIR}" ]; then
       OSD_J="${JOURNAL_DIR}/journal.${OSD_ID}"
       chown -R ceph. ${JOURNAL_DIR}
    else
       if [ -n "${JOURNAL}" ]; then
          OSD_J=${JOURNAL}
          chown -R ceph. $(dirname ${JOURNAL_DIR})
       else
          OSD_J=${OSD_PATH}/journal
       fi
    fi
    # log osd filesystem type
    FS_TYPE=`stat --file-system -c "%T" ${OSD_PATH}`
    log "OSD $OSD_PATH filesystem type: $FS_TYPE"
    echo "${CLUSTER}-${OSD_ID}: /usr/bin/ceph-osd ${CLI_OPTS[*]} -f -i ${OSD_ID} --osd-journal ${OSD_J} -k $OSD_KEYRING" | tee -a /etc/forego/"${CLUSTER}"/Procfile
  done
  log "SUCCESS"
  start_forego
}
