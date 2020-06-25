// Create WebSocket connection.
const socket = new WebSocket('ws://localhost:8080/ws/555/Anwar');

// Connection opened
socket.addEventListener('open', function (event) {
    socket.send('Hello this is Malcolm!');
});

// Listen for messages
socket.addEventListener('message', function (event) {
    console.log('Message from server ', event.data);
});