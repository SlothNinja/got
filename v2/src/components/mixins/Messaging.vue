<script>

import firebase from "firebase/app"
import "firebase/messaging"

const _ = require('lodash')

export default {
  data () {
    return {
      token: null,
      firebaseApp: null,
      firebaseMessaging: null
    }
  },
  methods: {
    getToken: function(f) {
      let self = this
      if (!self.fbmSupported) {
        console.log('Firebase messaging not supported')
        return
      }

      if (!_.isNull(self.token)) {
        if (_.isFunction(f)) {
          f(self.token)
        }
        return
      }

      self.fbMsg.getToken({ vapidKey: 'BLdpB0yUNJh2ZqJddGunLJ9oo_PSLkflgZzCBnqGCWMaPuId0YgY8woKilBstNmxGY7vk6s6lK3ecQQ-iTeXLVg' })
        .then((currentToken) => {
          self.token = currentToken
          if (_.isFunction(f)) {
            f(currentToken)
          }
        })
    }
  },
  computed: {
    fbApp: function () {
      let self = this
      if (!self.fbmSupported) {
        return null
      }
      if (_.isNull(self.firebaseApp)) {
        self.firebaseApp = firebase.initializeApp(self.fbmConfig)
      }
      return self.firebaseApp
    },
    fbMsg: function () {
      let self = this
      if (!self.fbmSupported) {
        return null
      }
      if (_.isNull(self.firebaseMessaging)) {
        self.firebaseMessaging = firebase.messaging(self.fbApp)
      }
      return self.firebaseMessaging
    },
    fbmSupported: function () {
      return firebase.messaging.isSupported()
    },
    fbmConfig: function () {
      return {
        apiKey: "AIzaSyDkTvPKy-TJQYo1P2yqScveK86Rf-mqblE",
        authDomain: "got-slothninja-games.firebaseapp.com",
        projectId: "got-slothninja-games",
        storageBucket: "got-slothninja-games.appspot.com",
        messagingSenderId: "623888087074",
        appId: "1:623888087074:web:9297f4c964c2f4726cf27b",
        measurementId: "G-QBX9QG6NH5"
      }
    }
  }
}
</script>
