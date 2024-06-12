# Websocket relay
Sometimes you want to be able to proxy a websocket, for example Whatsapp now blocks known cloud ip addresses for linking new devices.

This utility hosted on relay-ws.whatlist.io provides you a way of relaying a websocket proxy from your browser.

```
wm-server <=> relay-server <=> browser <=> web.whatsapp.com

wsocket1: browser <=> relay-server             (/registerUid?uid=u1, bidirectionally connected to wsocket2)
wsocket2: browser <=> web.whatsapp.com         (new websocket, bidirectionally connected to wsocket1)

// QR code scanning flow:
browser => whatlist-backend => wm-server <=> relay-server  (/relayExistingUid)
```
