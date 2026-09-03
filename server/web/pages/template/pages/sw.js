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
    "revision": "91da78d975999719cf82dc753eaa9635"
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
    "url": "static/vendor-CSj1mVf7.js",
    "revision": null
  }, {
    "url": "static/useTranslation-Bb3TMUoo.js",
    "revision": null
  }, {
    "url": "static/useTorrentDetail-BmJ4BtUY.js",
    "revision": null
  }, {
    "url": "static/torrsLink-Dv5wuOWf.js",
    "revision": null
  }, {
    "url": "static/torrents-BxJigCdd.js",
    "revision": null
  }, {
    "url": "static/torrentHelpers-Bl6IG-aK.js",
    "revision": null
  }, {
    "url": "static/runtime-BE2km-eZ.js",
    "revision": null
  }, {
    "url": "static/rolldown-runtime-8BhlS34s.js",
    "revision": null
  }, {
    "url": "static/maximize-2-DkfedEZJ.js",
    "revision": null
  }, {
    "url": "static/localPrefs-bn9onfF_.js",
    "revision": null
  }, {
    "url": "static/index-DWr6IVjV.js",
    "revision": null
  }, {
    "url": "static/index-B5ed88dO.css",
    "revision": null
  }, {
    "url": "static/hosts-CdUK6JJa.js",
    "revision": null
  }, {
    "url": "static/hls-DjLxvS_K.js",
    "revision": null
  }, {
    "url": "static/heroui-D1M9PDQJ.js",
    "revision": null
  }, {
    "url": "static/heart-BQDGfcus.js",
    "revision": null
  }, {
    "url": "static/gauge-CLHqP8Dl.js",
    "revision": null
  }, {
    "url": "static/film-hmCNXkUj.js",
    "revision": null
  }, {
    "url": "static/ellipsis-BHKU-GRs.js",
    "revision": null
  }, {
    "url": "static/createLucideIcon-DrUCZSM9.js",
    "revision": null
  }, {
    "url": "static/clapperboard-BcFCYMAz.js",
    "revision": null
  }, {
    "url": "static/circle-alert-Bi391bJo.js",
    "revision": null
  }, {
    "url": "static/authCredentials-CxJGlR23.js",
    "revision": null
  }, {
    "url": "static/VideoPlayer-a5DcaBDM.js",
    "revision": null
  }, {
    "url": "static/UnsafeButton-DBThzTKa.js",
    "revision": null
  }, {
    "url": "static/SettingsDialog-CuM4F_BD.js",
    "revision": null
  }, {
    "url": "static/ServerStatusDialog-eZJcEKJg.js",
    "revision": null
  }, {
    "url": "static/SearchDialog-B_f0WMGF.js",
    "revision": null
  }, {
    "url": "static/RemoveAllDialog-CS6D0l06.js",
    "revision": null
  }, {
    "url": "static/PosterPicker-CIynTTvz.js",
    "revision": null
  }, {
    "url": "static/PWAInstallationGuide-B7G2j7wa.js",
    "revision": null
  }, {
    "url": "static/MultiAddDialog-DZQdI0E7.js",
    "revision": null
  }, {
    "url": "static/ModalOpenContext-BvYJ7R-Z.js",
    "revision": null
  }, {
    "url": "static/ImportLibraryDialog-C0f9OjCU.js",
    "revision": null
  }, {
    "url": "static/ExportLibraryDialog-Bdiq6Pk6.js",
    "revision": null
  }, {
    "url": "static/EditTorrentDialog-B4hYzeHv.js",
    "revision": null
  }, {
    "url": "static/DonateSnackbar-Bg4vtYiC.js",
    "revision": null
  }, {
    "url": "static/DonateDialog-DPjjiiyB.js",
    "revision": null
  }, {
    "url": "static/DetailsDialog-DaMTS26M.js",
    "revision": null
  }, {
    "url": "static/CommandPalette-CVxO97FZ.js",
    "revision": null
  }, {
    "url": "static/CloseServerDialog-C5IguToQ.js",
    "revision": null
  }, {
    "url": "static/CategoriesDrawer-BB-lRdit.js",
    "revision": null
  }, {
    "url": "static/AppDialog-COlQUiDW.js",
    "revision": null
  }, {
    "url": "static/AndroidInstallBanner-BB8KanBb.js",
    "revision": null
  }, {
    "url": "static/AddDialog-DPR095uE.js",
    "revision": null
  }, {
    "url": "static/AboutDialog-DU1vEBDj.js",
    "revision": null
  }], {});
  workbox.cleanupOutdatedCaches();
  workbox.registerRoute(new workbox.NavigationRoute(workbox.createHandlerBoundToURL("index.html"), {
    denylist: [/^\/stream/, /^\/torrent/, /^\/torrents/, /^\/cache/, /^\/settings/, /^\/echo/, /^\/gst/, /^\/ffp/, /^\/download/, /^\/viewed/, /^\/search/, /^\/tmdb/, /^\/torznab/, /^\/storage/, /^\/waf/, /^\/mcp/, /^\/shutdown/, /^\/playlistall/, /^\/swagger/, /^\/stat/, /^\/magnets/]
  }));

}));
