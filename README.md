# Websocket relay
Sometimes you want to be able to proxy a websocket, for example Whatsapp now blocks known cloud ip addresses for linking new devices.

This utility hosted on relay-ws.whatlist.io provides you a way of relaying a websocket proxy from your browser.



wm-server <=> wm-proxy-server <=> browser <=> web.whatsapp.com

browser <=> web.whatsapp.com   (new websocket with uid=sid1)
browser <=> wm-proxy-server    (/registerUid)
browser calls whatlist-backend with whatsappWSSUrl=ws://wm-proxy-server/relayExistingUid?uid=sid1
wm-server <=> wm-proxy-server  (/relayExistingUid)
