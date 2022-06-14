{
    "watches": [
        {
            "name": "watch_http_404_error",
            "body": {
                "trigger": {
                    "schedule": {
                        "interval": "15m"
                    }
                },
                "input": {
                    "search": {
                        "request": {
                            "indices": [
                              {{ .Values.watcher.indices }}
                            ],
                            "body": {
                                "query": {
                                    "bool": {
                                        "must": {
                                            "match": {
                                                "response": 404
                                            }
                                        },
                                        "filter": {
                                            "range": {
                                                "@timestamp": {
                                                    "from": "{{`{{ctx.trigger.scheduled_time}}`}}||-5m",
                                                    "to": "{{`{{ctx.trigger.triggered_time}}`}}"
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                },
                "condition": {
                    "compare": {
                        "ctx.payload.hits.total": {
                            "gt": 0
                        }
                    }
                },
                "actions": {
                    "teams_webhook": {
                        "webhook": {
                            "scheme": "https",
                            "host": "outlook.office.com",
                            "port": 443,
                            "method": "post",
                            "path": "{{ .Values.watcher.webhooks.teams }}",
                            "params": {},
                            "headers": {},
                            "body": "{\"text\":\"HTTP 404 Errors\",\"fields\":[{\"title\":\"Count\",\"value\":\"{{`{{ctx.payload.hits.total}}`}}\",\"short\":true}]}]}"
                        }
                    }
                }
            }
        }
    ]
}
