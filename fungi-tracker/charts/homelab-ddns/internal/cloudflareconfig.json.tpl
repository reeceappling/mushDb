{
  "cloudflare": [
    {
      "authentication": {
        "api_token": {{- quote .Values.cloudflareApiToken }}
      },
      "zone_id": {{- quote .Values.cloudflareZoneId }},
      "subdomains": [
        {{- $nDomains := (len .Values.cloudflareSubdomains)}}
        {{- range $index, $subdomain := .Values.cloudflareSubdomains }}
        {{ printf "{%q:%q,%q:true}" "name" $subdomain "proxied"}}
        {{- if ( sub $nDomains $index | eq 1 ) }}{{- else }}{{- print ","}}{{ end }}
        {{- end }}
      ]
    }
  ],
  "a": true,
  "aaaa": false,
  "purgeUnknownRecords": false,
  "ttl": "Auto"
}