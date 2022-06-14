# Changelog

#### [v0.10.0](https://github.com/heptio/ark/releases/tag/v0.10.0) - 2018-11-15
  * [CHANGELOG-0.10.md][1]

#### [v0.9.11](https://github.com/heptio/ark/releases/tag/v0.9.11) - 2018-11-08

#### Bug Fixes
  * Fix bug preventing PV snapshots from being restored (#1040, @ncdc)

#### [v0.9.10](https://github.com/heptio/ark/releases/tag/v0.9.10) - 2018-11-01

#### Bug Fixes
  * restore storageclasses before pvs and pvcs (#594, @shubheksha)
  * AWS: Ensure that the order returned by ListObjects is consistent (#999, @bashofmann)
  * Add CRDs to list of prioritized resources (#424, @domenicrosati)
  * Verify PV doesn't exist before creating new volume (#609, @nrb)
  * Update README.md - Grammar mistake corrected (#1018, @midhunbiju)

#### [v0.9.9](https://github.com/heptio/ark/releases/tag/v0.9.9) - 2018-10-24

#### Bug Fixes
  * Check if initContainers key exists before attempting to remove volume mounts. (#927, @skriss)

#### [v0.9.8](https://github.com/heptio/ark/releases/tag/v0.9.8) - 2018-10-18

#### Bug Fixes
  * Discard service account token volume mounts from init containers on restore (#910, @james-powis)
  * Support --include-cluster-resources flag when creating schedule (#942, @captjt)
  * Remove logic to get a GCP project (#926, @shubheksha)
  * Only try to back up PVCs linked PV if the PVC's phase is Bound (#920, @skriss)
  * Claim ownership of new AWS volumes on Kubernetes cluster being restored into (#801, @ljakimczuk)
  * Remove timeout check when taking snapshots (#928, @carlisia)

#### [v0.9.7](https://github.com/heptio/ark/releases/tag/v0.9.7) - 2018-10-04

#### Bug Fixes
  * Preserve explicitly-specified node ports during restore (#712, @timoreimann)
  * Enable restoring resources with ownerReference set (#837, @mwieczorek)
  * Fix error when restoring ExternalName services (#869, @shubheksha)
  * remove restore log helper for accurate line numbers (#891, @skriss)
  * Display backup StartTimestamp in `ark backup get` output (#894, @marctc)
  * Fix restic restores when using namespace mappings (#900, @skriss)

#### [v0.9.6](https://github.com/heptio/ark/releases/tag/v0.9.6) - 2018-09-21

#### Bug Fixes
  * Discard service account tokens from non-default service accounts on restore (#843, @james-powis)
  * Update Docker images to use `alpine:3.8` (#852, @nrb)

#### [v0.9.5](https://github.com/heptio/ark/releases/tag/v0.9.5) - 2018-09-17

#### Bug Fixes
  * Fix issue causing restic restores not to work (#834, @skriss)

#### [v0.9.4](https://github.com/heptio/ark/releases/tag/v0.9.4) - 2018-09-05

#### Bug Fixes
  * Terminate plugin clients to resolve memory leaks (#797, @skriss)
  * Fix nil map errors when merging annotations (#812, @nrb)

#### [v0.9.3](https://github.com/heptio/ark/releases/tag/v0.9.3) - 2018-08-10

#### Bug Fixes
  * Initalize Prometheus metrics when creating a new schedule (#689, @lemaral)

#### [v0.9.2](https://github.com/heptio/ark/releases/tag/v0.9.2) - 2018-07-26

##### Bug Fixes:
  * Fix issue where modifications made by backup item actions were not being saved to backup tarball (#704, @skriss)

#### [v0.9.1](https://github.com/heptio/ark/releases/tag/v0.9.1) - 2018-07-23

##### Bug Fixes:
  * Require namespace for Ark's CRDs to already exist at server startup (#676, @skriss)
  * Require all Ark CRDs to exist at server startup (#683, @skriss)
  * Fix `latest` tagging in Makefile (#690, @skriss)
  * Make Ark compatible with clusters that don't have the `rbac.authorization.k8s.io/v1` API group (#682, @nrb)
  * Don't consider missing snapshots an error during backup deletion, limit backup deletion requests per backup to 1 (#687, @skriss)

#### [v0.9.0](https://github.com/heptio/ark/releases/tag/v0.9.0) - 2018-07-06

##### Highlights:
  * Ark now has support for backing up and restoring Kubernetes volumes using a free open-source backup tool called [restic](https://github.com/restic/restic).
    This provides users an out-of-the-box solution for backing up and restoring almost any type of Kubernetes volume, whether or not it has snapshot support
    integrated with Ark. For more information, see the [documentation](https://github.com/heptio/ark/blob/master/docs/restic.md).
  * Support for Prometheus metrics has been added! View total number of backup attempts (including success or failure), total backup size in bytes, and backup
    durations. More metrics coming in future releases!

##### All New Features:
  * Add restic support (#508 #532 #533 #534 #535 #537 #540 #541 #545 #546 #547 #548 #555 #557 #561 #563 #569 #570 #571 #606 #608 #610 #621 #631 #636, @skriss)
  * Add prometheus metrics (#531 #551 #564, @ashish-amarnath @nrb)
  * When backing up a service account, include cluster roles/cluster role bindings that reference it (#470, @skriss)
  * When restoring service accounts, copy secrets/image pull secrets into the target cluster even if the service account already exists (#403, @nrb)

##### Bug Fixes / Other Changes:
  * Upgrade to Kubernetes 1.10 dependencies (#417, @skriss)
  * Upgrade to go 1.10 and alpine 3.7 (#456, @skriss)
  * Display no excluded resources/namespaces as `<none>` rather than `*` (#453, @nrb)
  * Skip completed jobs and pods when restoring (#463, @nrb)
  * Set namespace correctly when syncing backups from object storage (#472, @skriss)
  * When building on macOS, bind-mount volumes with delegated config (#478, @skriss)
  * Add replica sets and daemonsets to cohabitating resources so they're not backed up twice (#482 #485, @skriss)
  * Shut down the Ark server gracefully on SIGINT/SIGTERM (#483, @skriss)
  * Only back up resources that support GET and DELETE in addition to LIST and CREATE (#486, @nrb)
  * Show a better error message when trying to get an incomplete restore's logs (#496, @nrb)
  * Stop processing when setting a backup deletion request's phase to `Deleting` fails (#500, @nrb)
  * Add library code to install Ark's server components (#437 #506, @marpaia)
  * Properly handle errors when backing up additional items (#512, @carlpett)
  * Run post hooks even if backup actions fail (#514, @carlpett)
  * GCP: fail backup if upload to object storage fails (#510, @nrb)
  * AWS: don't require `region` as part of backup storage provider config (#455, @skriss)
  * Ignore terminating resources while doing a backup (#526, @yastij)
  * Log to stdout instead of stderr (#553, @ncdc)
  * Move sample minio deployment's config to an emptyDir (#566, @runyontr)
  * Add `omitempty` tag to optional API fields (@580, @nikhita)
  * Don't restore PVs with a reclaim policy of `Delete` and no snapshot (#613, @ncdc)
  * Don't restore mirror pods (#619, @ncdc)

##### Docs Contributors:
  * @gianrubio
  * @castrojo
  * @dhananjaysathe
  * @c-knowles
  * @mattkelly
  * @ae-v
  * @hamidzr


#### [v0.8.3](https://github.com/heptio/ark/releases/tag/v0.8.3) - 2018-06-29

##### Bug Fixes:
  * Don't restore backup and restore resources to avoid possible data corruption (#622, @ncdc)

#### [v0.8.2](https://github.com/heptio/ark/releases/tag/v0.8.2) - 2018-06-01

##### Bug Fixes:
  * Don't crash when a persistent volume claim is missing spec.volumeName (#520, @ncdc)

#### [v0.8.1](https://github.com/heptio/ark/releases/tag/v0.8.1) - 2018-04-23

##### Bug Fixes:
  * Azure: allow pre-v0.8.0 backups with disk snapshots to be restored and deleted (#446 #449, @skriss)

#### [v0.8.0](https://github.com/heptio/ark/releases/tag/v0.8.0) - 2018-04-19

##### Highlights:
  * Backup deletion has been completely revamped to make it simpler and less error-prone. As a user, you still use the `ark backup delete` command to request deletion of a backup and its associated cloud
  resources; behind the scenes, we've switched to using a new `DeleteBackupRequest` Custom Resource and associated controller for processing deletion requests.
  * We've reduced the number of required fields in the Ark config. For Azure, `location` is no longer required, and for GCP, `project` is not needed.
  * Ark now copies tags from volumes to snapshots during backup, and from snapshots to new volumes during restore. 

##### Breaking Changes:
  * Ark has moved back to a single namespace (`heptio-ark` by default) as part of #383.

##### All New Features:
  * Add global `--kubecontext` flag to Ark CLI (#296, @blakebarnett)
  * Azure: support cross-resource group restores of volumes (#356 #378, @skriss)
  * AWS/Azure/GCP: copy tags from volumes to snapshots, and from snapshots to volumes (#341, @skriss)
  * Replace finalizer for backup deletion with `DeleteBackupRequest` custom resource & controller (#383 #431, @ncdc @nrb)
  * Don't log warnings during restore if an identical object already exists in the cluster (#405, @nrb)
  * Add bash & zsh completion support (#384, @containscafeine)
  
##### Bug Fixes / Other Changes:
  * Error from the Ark CLI if attempting to restore a non-existent backup (#302, @ncdc)
  * Enable running the Ark server locally for development purposes (#334, @ncdc)
  * Add examples to `ark schedule create` documentation (#331, @lypht)
  * GCP: Remove `project` requirement from Ark config (#345, @skriss)
  * Add `--from-backup` flag to `ark restore create` and allow custom restore names (#342 #409, @skriss)
  * Azure: remove `location` requirement from Ark config (#344, @skriss)
  * Add documentation/examples for storing backups in IBM Cloud Object Storage (#321, @roytman)
  * Reduce verbosity of hooks logging (#362, @skriss)
  * AWS: Add minimal IAM policy to documentation (#363 #419, @hopkinsth)
  * Don't restore events (#374, @sanketjpatel)
  * Azure: reduce API polling interval from 60s to 5s (#359, @skriss)
  * Switch from hostPath to emptyDir volume type for minio example (#386, @containscafeine)
  * Add limit ranges as a prioritized resource for restores (#392, @containscafeine)
  * AWS: Add documentation on using Ark with kube2iam (#402, @domderen)
  * Azure: add node selector so Ark pod is scheduled on a linux node (#415, @ffd2subroutine)
  * Error from the Ark CLI if attempting to get logs for a non-existent restore (#391, @containscafeine)
  * GCP: Add minimal IAM policy to documentation (#429, @skriss @jody-frankowski)

##### Upgrading from v0.7.1:
  Ark v0.7.1 moved the Ark server deployment into a separate namespace, `heptio-ark-server`. As of v0.8.0 we've
  returned to a single namespace, `heptio-ark`, for all Ark-related resources. If you're currently running v0.7.1,
  here are the steps you can take to upgrade:

1. Execute the steps from the **Credentials and configuration** section for your cloud:
    * [AWS](https://heptio.github.io/ark/v0.8.0/aws-config#credentials-and-configuration)
    * [Azure](https://heptio.github.io/ark/v0.8.0/azure-config#credentials-and-configuration)
    * [GCP](https://heptio.github.io/ark/v0.8.0/gcp-config#credentials-and-configuration)

    When you get to the secret creation step, if you don't have your `credentials-ark` file handy, 
    you can copy the existing secret from your `heptio-ark-server` namespace into the `heptio-ark` namespace:
    ```bash
    kubectl get secret/cloud-credentials -n heptio-ark-server --export -o json | \
      jq '.metadata.namespace="heptio-ark"' | \
      kubectl apply -f -
    ```

2. You can now safely delete the `heptio-ark-server` namespace:
    ```bash
    kubectl delete namespace heptio-ark-server
    ```

3. Execute the commands from the **Start the server** section for your cloud:
    * [AWS](https://heptio.github.io/ark/v0.8.0/aws-config#start-the-server)
    * [Azure](https://heptio.github.io/ark/v0.8.0/azure-config#start-the-server)
    * [GCP](https://heptio.github.io/ark/v0.8.0/gcp-config#start-the-server)


#### [v0.7.1](https://github.com/heptio/ark/releases/tag/v0.7.1) - 2018-02-22

Bug Fixes:
  * Run the Ark server in its own namespace, separate from backups/schedules/restores/config (#322, @ncdc)

#### [v0.7.0](https://github.com/heptio/ark/releases/tag/v0.7.0) - 2018-02-15

New Features:
  * Run the Ark server in any namespace (#272, @ncdc)
  * Add ability to delete backups and their associated data (#252, @skriss)
  * Support both pre and post backup hooks (#243, @ncdc)

Bug Fixes / Other Changes:
  * Switch from Update() to Patch() when updating Ark resources (#241, @skriss)
  * Don't fail the backup if a PVC is not bound to a PV (#256, @skriss)
  * Restore serviceaccounts prior to workload controllers (#258, @ncdc)
  * Stop removing annotations from PVs when restoring them (#263, @skriss)
  * Update GCP client libraries (#249, @skriss)
  * Clarify backup and restore creation messages (#270, @nrb)
  * Update S3 bucket creation docs for us-east-1 (#285, @lypht)

#### [v0.6.0](https://github.com/heptio/ark/tree/v0.6.0) - 2017-11-30

Highlights:
  * **Plugins** - We now support user-defined plugins that can extend Ark functionality to meet your custom backup/restore needs without needing to be compiled into the core binary. We support pluggable block and object stores as well as per-item backup and restore actions that can execute arbitrary logic, including modifying the items being backed up or restored. For more information see the [documentation](docs/plugins.md), which includes a reference to a fully-functional sample plugin repository. (#174 #188 #206 #213 #215 #217 #223 #226)
  * **Describers** - The Ark CLI now includes `describe` commands for `backups`, `restores`, and `schedules` that provide human-friendly representations of the relevant API objects.

Breaking Changes:
  * The config object format has changed. In order to upgrade to v0.6.0, the config object will have to be updated to match the new format. See the [examples](examples) and [documentation](docs/config-definition.md) for more information.
  * The restore object format has changed. The `warnings` and `errors` fields are now ints containing the counts, while full warnings and errors are now stored in the object store instead of etcd. Restore objects created prior to v.0.6.0 should be deleted, or a new bucket used, and the old restore objects deleted from Kubernetes (`kubectl -n heptio-ark delete restore --all`).

All New Features:
  * Add `ark plugin add` and `ark plugin remove` commands #217, @skriss
  * Add plugin support for block/object stores, backup/restore item actions #174 #188 #206 #213 #215 #223 #226, @skriss @ncdc
  * Improve Azure deployment instructions #216, @ncdc
  * Change default TTL for backups to 30 days #204, @nrb
  * Improve logging for backups and restores #199, @ncdc
  * Add `ark backup describe`, `ark schedule describe` #196, @ncdc
  * Add `ark restore describe` and move restore warnings/errors to object storage #173 #201 #202, @ncdc
  * Upgrade to client-go v5.0.1, kubernetes v1.8.2 #157, @ncdc
  * Add Travis CI support #165 #166, @ncdc

Bug Fixes:
  * Fix log location hook prefix stripping #222, @ncdc
  * When running `ark backup download`, remove file if there's an error #154, @ncdc
  * Update documentation for AWS KMS Key alias support #163, @lli-hiya
  * Remove clock from `volume_snapshot_action` #137, @athampy

#### [v0.5.1](https://github.com/heptio/ark/tree/v0.5.1) - 2017-11-06
Bug fixes:
  * If a Service is headless, retain ClusterIP = None when backing up and restoring.
  * Use the specifed --label-selector when listing backups, schedules, and restores.
  * Restore namespace mapping functionality that was accidentally broken in 0.5.0.
  * Always include namespaces in the backup, regardless of the --include-cluster-resources setting.

#### [v0.5.0](https://github.com/heptio/ark/tree/v0.5.0) - 2017-10-26
Breaking changes:
  * The backup tar file format has changed. Backups created using previous versions of Ark cannot be restored using v0.5.0.
  * When backing up one or more specific namespaces, cluster-scoped resources are no longer backed up by default, with the exception of PVs that are used within the target namespace(s). Cluster-scoped resources can still be included by explicitly specifying `--include-cluster-resources`.

New features:
  * Add customized user-agent string for Ark CLI
  * Switch from glog to logrus
  * Exclude nodes from restoration
  * Add a FAQ
  * Record PV availability zone and use it when restoring volumes from snapshots
  * Back up the PV associated with a PVC
  * Add `--include-cluster-resources` flag to `ark backup create`
  * Add `--include-cluster-resources` flag to `ark restore create`
  * Properly support resource restore priorities across cluster-scoped and namespace-scoped resources
  * Support `ark create ...` and `ark get ...`
  * Make ark run as cluster-admin
  * Add pod exec backup hooks
  * Support cross-compilation & upgrade to go 1.9
  
Bug fixes:
  * Make config change detection more robust

#### [v0.4.0](https://github.com/heptio/ark/tree/v0.4.0) - 2017-09-14
Breaking changes:
  * Snapshotting and restoring volumes is now enabled by default
  * The --namespaces flag for 'ark restore create' has been replaced by --include-namespaces and
    --exclude-namespaces

New features:
  * Support for S3 SSE with KMS
  * Cloud provider configurations are validated at startup
  * The persistentVolumeProvider is now optional
  * Restore objects are garbage collected
  * Each backup now has an associated log file, viewable via 'ark backup logs'
  * Each restore now has an associated log file, viewable via 'ark restore logs'
  * Add --include-resources/--exclude-resources for restores

Bug fixes:
  * Only save/use iops for io1 volumes on AWS
  * When restoring, try to retrieve the Backup directly from object storage if it's not found
  * When syncing Backups from object storage to Kubernetes, don't return at the first error
    encountered
  * More closely match how kubectl performs kubeconfig resolution
  * Increase default Azure API request timeout to 2 minutes
  * Update Azure diskURI to match diskName

#### [v0.3.3](https://github.com/heptio/ark/tree/v0.3.3) - 2017-08-10
  * Treat the first field in a schedule's cron expression as minutes, not seconds

#### [v0.3.2](https://github.com/heptio/ark/tree/v0.3.2) - 2017-08-07
  * Add client-go auth provider plugins for Azure, GCP, OIDC

#### [v0.3.1](https://github.com/heptio/ark/tree/v0.3.1) - 2017-08-03
  * Fix Makefile VERSION

#### [v0.3.0](https://github.com/heptio/ark/tree/v0.3.0) - 2017-08-03
  * Initial Release

[1]: https://github.com/heptio/ark/blob/master/changelogs/CHANGELOG-0.10.md
