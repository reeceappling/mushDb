{
  "cloudflare": [
    {
      "authentication": {
        "api_token": {{- quote .Values.cloudflareApiToken }}
      },
      "zone_id":, {{ .Values.cloudflareZoneId }},
      "subdomains": [
        {{- with .Values.cloudflareSubdomains }}
        {{- range $index, $subdomain := .Values.cloudflareSubdomains}}
        {{ if $index ne 0 -}}{{- printf "," }}{{- end }}}{{- printf "{%q:%q,%q:true}" "name" $subdomain "proxied"}}
        {{- end }}
        {{- end}}
      ]
    }
  ],
  "a": true,
  "aaaa": false,
  "purgeUnknownRecords": false,
  "ttl": "Auto"
}