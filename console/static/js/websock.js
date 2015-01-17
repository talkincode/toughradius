var wsock = {
    socket: null,
    timer: null,

    start: function (addr,port) {
        var url = "ws://"+addr+":"+port;
        wsock.socket = new WebSocket(url);
        wsock.socket.onmessage = function (event) {
            wsock.onMessage(JSON.parse(event.data));
        }
        wsock.socket.onclose = wsock.onClose;
        wsock.socket.onopen = wsock.onOpen;
    },

    sendMessage: function(message){
        data = JSON.stringify(message);
        wsock.socket.send(data);
    },

    onMessage: function (message) {

    },

    onOpen: function(event){

    },

    onClose: function(event){
        wsock.start();
    }
};