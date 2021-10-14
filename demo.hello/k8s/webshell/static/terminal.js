// Includes terminal open, close and io functions.
const ws_host = 'localhost:8090'
let term, ws_conn

function connectWS(namespace, pod, container_name) {
  // for debug: file:///local_path/to/terminal.html?namespace=mini-test-ns&pod=hello-minikube-59ddd8676b-vkl26
  // let namespace = getQueryVariable('namespace')
  // let pod = getQueryVariable('pod')
  // let container_name = getQueryVariable('container_name')

  if (!Boolean(container_name)) {
    container_name = 'null'
  }
  if (!Boolean(namespace) || !Boolean(pod)) {
    alert('Namespace or pod is empty in query!')
    return
  }
  console.log(`ns: ${namespace}, pod: ${pod}, container: ${container_name}`)

  initTerminal()
  initWebsocket(namespace, pod, container_name)
}

function initTerminal() {
  if (term) {
    return
  }

  if (!window['WebSocket']) {
    let item = document.getElementById('terminal')
    item.innerHTML = '<h2>Your browser does not support WebSockets.</h2>'
    return
  }

  term = new Terminal({
    'cursorBlink': true,
  })
  term.open(document.getElementById('terminal'))
  term.writeln(`Terminal is started at ${new Date()}`)
  term.fit()
  // term.toggleFullScreen(true)

  // send req data to backend websocket
  term.on('data', function(data) {
    msg = { operation: 'stdin', data: data }
    ws_conn.send(JSON.stringify(msg))
  })

  term.on('resize', function(size) {
    console.log('Term resize:', size)
    msg = { operation: 'resize', cols: size.cols, rows: rows }
    ws_conn.send(JSON.stringify(msg))
  })
}

function initWebsocket(namespace, pod, container) {
  if (Boolean(ws_conn)) {
    ws_conn.close()
  }

  let url = `ws://${ws_host}/ws/${namespace}/${pod}/${container}/webshell`
  console.log(`Connect to ws url: ${url}`)
  ws_conn = new WebSocket(url)

  ws_conn.onopen = function(e) {
    term.writeln(`Connecting to pod [ ${pod} ] container [ ${container} ]...`)
    term.write('\r')
    msg = { operation: 'stdin', data: 'export TERM=xterm\r' }
    ws_conn.send(JSON.stringify(msg))
    // term.clear()
  }

  // write resp data to term
  ws_conn.onmessage = function(event) {
    msg = JSON.parse(event.data)
    if (msg.operation === 'stdout') {
      term.write(msg.data)
    } else {
      console.warn('Invalid msg operation:', msg)
    }
  }

  ws_conn.onclose = function(event) {
    if (event.wasClean) {
      console.log(`[Close] Connection closed cleanly, code=${event.code} reason=${event.reason}`)
    } else {
      console.warn('[Close] Connection died')
      term.writeln("")
    }
    term.writeln('Connection Reset By Peer! Try Refresh.')
  }

  ws_conn.onerror = function(error) {
    console.error('[Error] Connection error')
    term.write("Error: " + error.message)
    term.destroy()
  }
}

function disconnectWS() {
  if (!Boolean(ws_conn)) {
    alert("WebSocket is already closed!")
  } else {
    ws_conn.close()
    ws_conn = null
  }
}

function getQueryVariable(variable) {
  let query = window.location.search.substring(1)
  let vars = query.split('&')

  for (let i = 0; i < vars.length; i++) {
    let pair = vars[i].split('=')
    if (pair[0] == variable) {
      return pair[1]
    }
  }
  return false
}
