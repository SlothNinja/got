module.exports = {
  "transpileDependencies": [
    "vuetify"
  ],
  pwa: {
    // disable: process.env.NODE_ENV === 'development',
    name: 'Guild of Thieves',
    themeColor: '#4DBA87',
    msTileColor: '#000000',
    appleMobileWebAppCapable: 'yes',
    appleMobileWebAppStatusBarStyle: 'black',

    // configure the workbox plugin
    workboxPluginMode: 'InjectManifest',
    workboxOptions: {
      swSrc: 'src/firebase-messaging-sw.js',
      //    swDest:'js/sw.js',
      //    importsDirectory: 'js/',
    }
  }
}
