input {
{{- range $environment, $connection := .Values.stunnel.connections }}
  redis {
    id => {{ $environment | quote }}
    host => {{ $connection.local.host | quote }}
    port => {{ $connection.local.port | quote }}
    batch_count => "5000"
    data_type => "list"
    key => "logstash"
    password => {{ $connection.redis.key | quote }}
    type => "logstash-input"
    threads => "20"
    tags => [{{ $environment | quote }}]
  }
{{- end }}
}

filter {
  # Example of filter
  # if [type] == "logstash-input" {
  #   if [logger_name] == "org.logger" and [event-name] == "event-name" {
  #     drop {}
  #   }
  # }

}

output {
{{- range $environment, $connection := .Values.stunnel.connections }}
  if {{ $environment | quote }} in [tags] {
    if [type] == "haproxy" {
      elasticsearch {
        hosts => ["elasticsearch:9200"]
        index => "{{ $environment }}-haproxy-%{+YYYY.MM.dd}"
        manage_template => false
        document_type => "haproxy"
      }
    } if [type] == "syslog" {
      elasticsearch {
        hosts => ["elasticsearch:9200"]
        index => "{{ $environment }}-syslog-%{+YYYY.MM.dd}"
        manage_template => false
        document_type => "syslog"
      }
    } else {
      elasticsearch {
        hosts => ["elasticsearch:9200"]
        index => "{{ $environment }}-logstash-%{+YYYY.MM.dd}"
        manage_template => false
        document_type => "logstash-input"
      }
  }
  }
{{- end }}
}