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
    "revision": "454c2bf70710f2097fea3672b3ec273e"
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
    "url": "static/vendor-Du-e_o69.js",
    "revision": null
  }, {
    "url": "static/useTranslation-BjgSEyAy.js",
    "revision": null
  }, {
    "url": "static/useTorrentDetail-BcbsyTWi.js",
    "revision": null
  }, {
    "url": "static/torrsLink-CrcAIGZk.js",
    "revision": null
  }, {
    "url": "static/torrents-Dyi4HRXE.js",
    "revision": null
  }, {
    "url": "static/torrentHelpers-DBgO9FM3.js",
    "revision": null
  }, {
    "url": "static/states-C1wEik_A.js",
    "revision": null
  }, {
    "url": "static/settings-CaNicdu0.js",
    "revision": null
  }, {
    "url": "static/runtime-mUYutevX.js",
    "revision": null
  }, {
    "url": "static/rolldown-runtime-8BhlS34s.js",
    "revision": null
  }, {
    "url": "static/maximize-2-bSMzmtyi.js",
    "revision": null
  }, {
    "url": "static/localPrefs-C6BvauQz.js",
    "revision": null
  }, {
    "url": "static/index-BTkR-k2X.css",
    "revision": null
  }, {
    "url": "static/index-B0YvqBaH.js",
    "revision": null
  }, {
    "url": "static/hosts-9tTxUrl2.js",
    "revision": null
  }, {
    "url": "static/hls-BPUeZg91.js",
    "revision": null
  }, {
    "url": "static/heroui-CRxKGtYz.js",
    "revision": null
  }, {
    "url": "static/heart-G6_5busT.js",
    "revision": null
  }, {
    "url": "static/gauge-Bn-mzinc.js",
    "revision": null
  }, {
    "url": "static/film-BAfZvQ73.js",
    "revision": null
  }, {
    "url": "static/ellipsis-zchaEk8m.js",
    "revision": null
  }, {
    "url": "static/createLucideIcon-tJym343k.js",
    "revision": null
  }, {
    "url": "static/clapperboard-B_72ALO_.js",
    "revision": null
  }, {
    "url": "static/circle-alert-D0dV3IfU.js",
    "revision": null
  }, {
    "url": "static/authCredentials-D2T-PHHc.js",
    "revision": null
  }, {
    "url": "static/VideoPlayer-D5rN-ZiQ.js",
    "revision": null
  }, {
    "url": "static/UnsafeButton-V7Ojnw5Q.js",
    "revision": null
  }, {
    "url": "static/SettingsDialog-1YwwZDwe.js",
    "revision": null
  }, {
    "url": "static/ServerStatusDialog-CUpm0dxg.js",
    "revision": null
  }, {
    "url": "static/SearchDialog-BXOir4hk.js",
    "revision": null
  }, {
    "url": "static/RemoveAllDialog-uxxUudpo.js",
    "revision": null
  }, {
    "url": "static/PosterPicker-DA6mXARy.js",
    "revision": null
  }, {
    "url": "static/PWAInstallationGuide-D62YgYok.js",
    "revision": null
  }, {
    "url": "static/MultiAddDialog-BSa9BNP2.js",
    "revision": null
  }, {
    "url": "static/ModalOpenContext-Bzn3py4o.js",
    "revision": null
  }, {
    "url": "static/ImportLibraryDialog-CgdchLkl.js",
    "revision": null
  }, {
    "url": "static/ExportLibraryDialog-CPTX-z3_.js",
    "revision": null
  }, {
    "url": "static/EditTorrentDialog-C00oJlfK.js",
    "revision": null
  }, {
    "url": "static/DonateSnackbar-BUcWYT4H.js",
    "revision": null
  }, {
    "url": "static/DonateDialog-CQtPsneR.js",
    "revision": null
  }, {
    "url": "static/DetailsDialog-BH6PPD8o.js",
    "revision": null
  }, {
    "url": "static/CommandPalette-BUYoIh0O.js",
    "revision": null
  }, {
    "url": "static/CloseServerDialog-EeZPHaFk.js",
    "revision": null
  }, {
    "url": "static/CategoriesDrawer-CXz87Io7.js",
    "revision": null
  }, {
    "url": "static/AppDialog-BzPllanF.js",
    "revision": null
  }, {
    "url": "static/AndroidInstallBanner-ESWvST2z.js",
    "revision": null
  }, {
    "url": "static/AddDialog-CGALvlrs.js",
    "revision": null
  }, {
    "url": "static/AboutDialog-DJQB0sxL.js",
    "revision": null
  }], {});
  workbox.cleanupOutdatedCaches();
  workbox.registerRoute(new workbox.NavigationRoute(workbox.createHandlerBoundToURL("index.html"), {
    denylist: [/^\/stream/, /^\/torrent/, /^\/torrents/, /^\/cache/, /^\/settings/, /^\/echo/, /^\/gst/, /^\/ffp/, /^\/download/, /^\/viewed/, /^\/search/, /^\/tmdb/, /^\/torznab/, /^\/storage/, /^\/waf/, /^\/mcp/, /^\/shutdown/, /^\/playlistall/, /^\/swagger/, /^\/stat/, /^\/magnets/]
  }));

}));
