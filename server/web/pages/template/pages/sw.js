/**
 * Copyright 2018 Google Inc. All Rights Reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// If the loader is already loaded, just stop.
if (!self.define) {
  let registry = {};

  // Used for `eval` and `importScripts` where we can't get script URL by other means.
  // In both cases, it's safe to use a global var because those functions are synchronous.
  let nextDefineUri;

  const singleRequire = (uri, parentUri) => {
    uri = new URL(uri + ".js", parentUri).href;
    return registry[uri] || (
      
        new Promise(resolve => {
          if ("document" in self) {
            const script = document.createElement("script");
            script.src = uri;
            script.onload = resolve;
            document.head.appendChild(script);
          } else {
            nextDefineUri = uri;
            importScripts(uri);
            resolve();
          }
        })
      
      .then(() => {
        let promise = registry[uri];
        if (!promise) {
          throw new Error(`Module ${uri} didn’t register its module`);
        }
        return promise;
      })
    );
  };

  self.define = (depsNames, factory) => {
    const uri = nextDefineUri || ("document" in self ? document.currentScript.src : "") || location.href;
    if (registry[uri]) {
      // Module is already loading or loaded.
      return;
    }
    let exports = {};
    const require = depUri => singleRequire(depUri, uri);
    const specialDeps = {
      module: { uri },
      exports,
      require
    };
    registry[uri] = Promise.all(depsNames.map(
      depName => specialDeps[depName] || require(depName)
    )).then(deps => {
      factory(...deps);
      return exports;
    });
  };
}
define(['./workbox-7e5eb42b'], (function (workbox) { 'use strict';

  self.skipWaiting();
  workbox.clientsClaim();
  /**
   * The precacheAndRoute() method efficiently caches and responds to
   * requests for URLs in the manifest.
   * See https://goo.gl/S9QRab
   */
  workbox.precacheAndRoute([{
    "url": "mstile-150x150.png",
    "revision": "5701d873784a3f27c3dffb8a15736079"
  }, {
    "url": "logo.png",
    "revision": "f5db7f222c6f454cdf40c0793ef529f3"
  }, {
    "url": "index.html",
    "revision": "c269b9a8dbff0c8a982f192a120d4d97"
  }, {
    "url": "icon.png",
    "revision": "4425f6f2b52d2deaf2374ff63c682bcc"
  }, {
    "url": "favicon.ico",
    "revision": "80c0375465582708d791bcab9b4053af"
  }, {
    "url": "favicon-32x32.png",
    "revision": "d34d53cbac4d729b0f49677fecbb48ac"
  }, {
    "url": "favicon-16x16.png",
    "revision": "1cabe79322b89571937189d95cbe968b"
  }, {
    "url": "dlnaicon-48.png",
    "revision": "94fb147303f8625b3f0dd99a385d72d8"
  }, {
    "url": "dlnaicon-120.png",
    "revision": "a741da374199bad465c9e2576da666d0"
  }, {
    "url": "static/workbox-window.prod.es5-Bd17z0YL.js",
    "revision": null
  }, {
    "url": "static/vendor-Di2vZqaS.js",
    "revision": null
  }, {
    "url": "static/useTranslation-DCdmzRKG.js",
    "revision": null
  }, {
    "url": "static/useTorrentDetail-DahPgSZj.js",
    "revision": null
  }, {
    "url": "static/torrsLink-Dv5wuOWf.js",
    "revision": null
  }, {
    "url": "static/torrents-DXgeeQn_.js",
    "revision": null
  }, {
    "url": "static/torrentHelpers-BQS4ZLds.js",
    "revision": null
  }, {
    "url": "static/states-C1wEik_A.js",
    "revision": null
  }, {
    "url": "static/settings-B0xlMFmP.js",
    "revision": null
  }, {
    "url": "static/runtime-CA4Q2Ve3.js",
    "revision": null
  }, {
    "url": "static/rolldown-runtime-8BhlS34s.js",
    "revision": null
  }, {
    "url": "static/maximize-2-65p__kna.js",
    "revision": null
  }, {
    "url": "static/localPrefs-bn9onfF_.js",
    "revision": null
  }, {
    "url": "static/index-Dajr2-Lf.js",
    "revision": null
  }, {
    "url": "static/index-BaaAu-Dp.css",
    "revision": null
  }, {
    "url": "static/hosts-CMo7Owjk.js",
    "revision": null
  }, {
    "url": "static/hls-DbOvCaf2.js",
    "revision": null
  }, {
    "url": "static/heroui-BRk0eEAP.js",
    "revision": null
  }, {
    "url": "static/heart-DD0hL22U.js",
    "revision": null
  }, {
    "url": "static/gauge-C_zd3x55.js",
    "revision": null
  }, {
    "url": "static/film-dmZnMgPQ.js",
    "revision": null
  }, {
    "url": "static/ellipsis-B1KO_-cg.js",
    "revision": null
  }, {
    "url": "static/createLucideIcon-D0SXGUxE.js",
    "revision": null
  }, {
    "url": "static/clapperboard-CfZq_Xq0.js",
    "revision": null
  }, {
    "url": "static/circle-alert-BruxOFtu.js",
    "revision": null
  }, {
    "url": "static/authCredentials-Re79LSk2.js",
    "revision": null
  }, {
    "url": "static/VideoPlayer-0PoDb3vl.js",
    "revision": null
  }, {
    "url": "static/UnsafeButton-T2aHaeOA.js",
    "revision": null
  }, {
    "url": "static/SettingsDialog-BgaHgUGq.js",
    "revision": null
  }, {
    "url": "static/ServerStatusDialog-B6gX4DYi.js",
    "revision": null
  }, {
    "url": "static/SearchDialog-Fp5xVp9u.js",
    "revision": null
  }, {
    "url": "static/RemoveAllDialog-D4h4ye83.js",
    "revision": null
  }, {
    "url": "static/PosterPicker-DTbVOjqd.js",
    "revision": null
  }, {
    "url": "static/PWAInstallationGuide-D6Y0RuQh.js",
    "revision": null
  }, {
    "url": "static/MultiAddDialog-RBUXn5rS.js",
    "revision": null
  }, {
    "url": "static/ModalOpenContext-D86G_T3u.js",
    "revision": null
  }, {
    "url": "static/ImportLibraryDialog-Dm-GldMu.js",
    "revision": null
  }, {
    "url": "static/ExportLibraryDialog-BJNU7LgM.js",
    "revision": null
  }, {
    "url": "static/EditTorrentDialog-Bc5KDusM.js",
    "revision": null
  }, {
    "url": "static/DonateSnackbar-DEpFConQ.js",
    "revision": null
  }, {
    "url": "static/DonateDialog-9JUdIy-I.js",
    "revision": null
  }, {
    "url": "static/DetailsDialog-DOHBIb3i.js",
    "revision": null
  }, {
    "url": "static/CommandPalette-BA2pQlN7.js",
    "revision": null
  }, {
    "url": "static/CloseServerDialog-BeuNdbCK.js",
    "revision": null
  }, {
    "url": "static/CategoriesDrawer-_PNJsCII.js",
    "revision": null
  }, {
    "url": "static/AppDialog-CtyGKA7K.js",
    "revision": null
  }, {
    "url": "static/AndroidInstallBanner-CJILf5hP.js",
    "revision": null
  }, {
    "url": "static/AddDialog-CO8i_3mp.js",
    "revision": null
  }, {
    "url": "static/AboutDialog-BdFClpR3.js",
    "revision": null
  }], {});
  workbox.cleanupOutdatedCaches();
  workbox.registerRoute(new workbox.NavigationRoute(workbox.createHandlerBoundToURL("index.html"), {
    denylist: [/^\/stream/, /^\/torrent/, /^\/torrents/, /^\/cache/, /^\/settings/, /^\/echo/, /^\/gst/, /^\/ffp/, /^\/download/, /^\/viewed/, /^\/search/, /^\/tmdb/, /^\/torznab/, /^\/storage/, /^\/waf/, /^\/mcp/, /^\/shutdown/, /^\/playlistall/, /^\/swagger/, /^\/stat/, /^\/magnets/]
  }));

}));
