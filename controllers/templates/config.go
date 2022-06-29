package templates

const CONFIG = `
tunnel: {{ .TunnelID }}
credentials-file: /root/.cloudflared/{{ .TunnelID }}.json
warp-routing:
  enabled: true
ingress:
  - service: {{ .Service }}
    originRequest:
      originServerName: {{ .Domain }}
`
