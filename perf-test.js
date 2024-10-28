/**
 * This script is a performance test for a WebSocket server using k6.
 * It establishes a connection to the server, subscribes to a chat room, sends a message,
 * listens for incoming messages, unsubscribes from the chat room, and closes the connection.
 *
 * The test is designed to simulate a number of users sending messages to each other.
 * The number of users is defined by the `options.stages` array.
 *
 * The test measures the time it takes to connect to the WebSocket server, send a message, and
 * receive a message. It also measures the number of messages sent and received, and the error
 * rate of the messages.
 */

import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Trend, Counter, Rate } from 'k6/metrics';

// WebSocket URL
const url = 'ws://localhost:8383/ws';

// Metrics for the performance analysis
const connectTime = new Trend('ws_connect_time', 'Time it takes to connect to the WebSocket server');
const sendMessageTime = new Trend('ws_send_message_time', 'Time it takes to send a message');
const messageReceivedTime = new Trend('ws_message_received_time', 'Time it takes to receive a message');
const messagesSentCounter = new Counter('ws_messages_sent', 'Number of messages sent');
const messagesReceivedCounter = new Counter('ws_messages_received', 'Number of messages received');
const messageErrorRate = new Rate('ws_message_error_rate', 'Rate of errors when receiving messages');

// Number of messages to send per user
const messagesToSend = 5;
const channels = ['chat'];

// Test options
export let options = {
    stages: [
        { duration: '2m', target: 1000  },  // Ramp-up to 1000 users over 2 minute
        { duration: '4m', target: 2800  },  // Ramp-up to 2800 users over 4 minutes
        { duration: '4m', target: 2800  },  // Sustain 1000 users for 2 minutes
        { duration: '2m', target: 0     },  // Ramp-down to 0 users
    ],
};

export default function () {
    const vuId = __VU;
    const params = { tags: { vuId: `VU-${vuId}` } };

    const res = ws.connect(url, params, function (socket) {
        let startConnectTime = new Date().getTime();

        socket.on('open', () => {
            // console.log(`[VU-${vuId}] Verbindung hergestellt`);
            connectTime.add(new Date().getTime() - startConnectTime);

            // Subscribe to the chat rooms
            subscribeToChannels(socket, channels, vuId);

            // Send messages
            sendMessages(socket, vuId);

            // Every second user will unsubscribe from the chat rooms and close the connection
            if (vuId % 2 !== 0) {
                // Unsubscribe from the chat rooms
                sleep(5);
                unsubscribeFromChannels(socket, channels, vuId);
                
                // Close the connection
                sleep(5);
                socket.close();
            }
        });

        socket.on('message', (msg) => {
            // Handle the message
            handleMessage(msg, vuId);
        });

        socket.on('close', () => {
            // console.log(`[VU-${vuId}] Verbindung beendet`);
        });

        socket.on('error', (e) => {
            console.error(`[VU-${vuId}] WebSocket-Fehler:`, e);
        });
    });

    // Check if the connection was successful
    check(res, {
        'Erfolgreich verbunden': (r) => r && r.status === 101,
    });

    sleep(1);
}

function subscribeToChannels(socket, channels, vuId) {
    channels.forEach(channel => {
        socket.send(JSON.stringify(['subscribe:' + channel, {}]));
        // console.log(`[VU-${vuId}] Abonniert: ${channel}`);
    });
}

function unsubscribeFromChannels(socket, channels, vuId) {
    channels.forEach(channel => {
        socket.send(JSON.stringify(['unsubscribe:' + channel, {}]));
        // console.log(`[VU-${vuId}] Deabonniert: ${channel}`);
    });
}

function sendMessages(socket, vuId) {
    for (let i = 0; i < messagesToSend; i++) {
        const startSendMessageTime = new Date().getTime();

        const message = {
            from: `user_${vuId}`,
            message: `text_${i}`
        };

        // Send the message and log it
        socket.send(JSON.stringify(['chat:message', message]));
        // console.log(`[VU-${vuId}] Gesendete Nachricht: ${JSON.stringify(message)}`);
        
        sendMessageTime.add(new Date().getTime() - startSendMessageTime);
        messagesSentCounter.add(1);

        // Random pause between the messages (2-14 seconds)
        sleep(Math.random() * 12 + 2);
    }
}

function handleMessage(msg, vuId) {
    const startMessageReceivedTime = new Date().getTime();

    const messageData = JSON.parse(msg);

    if (Array.isArray(messageData)) {
        const [type, payload] = messageData;

        if (type === 'chat:message') {
            // console.log(`[VU-${vuId}] Empfangene Nachricht: ${JSON.stringify(payload)}`);
            messageReceivedTime.add(new Date().getTime() - startMessageReceivedTime);
            messagesReceivedCounter.add(1);

            check(payload, {
                'Nachricht wurde empfangen': (m) => m.message !== '',
            });
        }
    } else {
        handleParsingError(msg, vuId);
    }
}

function handleParsingError(msg, vuId) {
    messageErrorRate.add(1);
    console.error(`[VU-${vuId}] Fehler beim Parsen der Nachricht: ${msg}`);
}

function handleProcessingError(error, vuId) {
    messageErrorRate.add(1);
    console.error(`[VU-${vuId}] Fehler beim Verarbeiten der Nachricht: ${error}`);
}
