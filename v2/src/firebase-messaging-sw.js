/**
 * Welcome to your Workbox-powered service worker!
 *
 * You'll need to register this file in your web app and you should
 * disable HTTP caching for this file too.
 * See https://goo.gl/nhQhGp
 *
 * The rest of the code is auto-generated. Please don't update this file
 * directly; instead, make changes to your Workbox build configuration
 * and re-run your build process.
 * See https://goo.gl/2aRDsh
 */

importScripts("https://storage.googleapis.com/workbox-cdn/releases/4.3.1/workbox-sw.js");

workbox.core.setCacheNameDetails({prefix: "got"});

importScripts('https://www.gstatic.com/firebasejs/8.3.0/firebase-app.js')
importScripts('https://www.gstatic.com/firebasejs/8.3.0/firebase-messaging.js')

if (firebase.messaging.isSupported()) {
firebase.initializeApp({
  apiKey: "AIzaSyDkTvPKy-TJQYo1P2yqScveK86Rf-mqblE",
  authDomain: "got-slothninja-games.firebaseapp.com",
  projectId: "got-slothninja-games",
  storageBucket: "got-slothninja-games.appspot.com",
  messagingSenderId: "623888087074",
  appId: "1:623888087074:web:9297f4c964c2f4726cf27b",
  measurementId: "G-QBX9QG6NH5"
})

// Retrieve an instance of Firebase Messaging so that it can handle background
// messages.
const messaging = firebase.messaging()
}
// messaging.setBackgroundMessageHandler(function (payload) {
//     const promiseChain = clients
//         .matchAll({
//             type: "window",
//             includeUncontrolled: true
//         })
//         .then(windowClients => {
//             for (let i = 0; i < windowClients.length; i++) {
//                 const windowClient = windowClients[i];
//                 windowClient.postMessage(payload);
//             }
//         })
//         .then(() => {
//             return registration.showNotification("New Message");
//         });
//     return promiseChain;
// });
// self.addEventListener('notificationclick', function (event) {
//     console.log('notification received: ', event)
// });
// self.addEventListener('message', (event) => {
//   if (event.data && event.data.type === 'SKIP_WAITING') {
//     self.skipWaiting();
//   }
// });

/**
 * The workboxSW.precacheAndRoute() method efficiently caches and responds to
 * requests for URLs in the manifest.
 * See https://goo.gl/S9QRab
 */
self.__precacheManifest = [].concat(self.__precacheManifest || []);
workbox.precaching.precacheAndRoute(self.__precacheManifest, {});
