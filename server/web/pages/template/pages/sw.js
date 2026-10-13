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
    "revision": "1751cd91b0f96ee94086eed94950b3bf"
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
    "url": "static/vendor-BLQlkSIf.js",
    "revision": null
  }, {
    "url": "static/useTranslation-BhLCZ20K.js",
    "revision": null
  }, {
    "url": "static/useTorrentDetail-CNUIxO9d.js",
    "revision": null
  }, {
    "url": "static/torrsLink-Dv5wuOWf.js",
    "revision": null
  }, {
    "url": "static/torrents-Dg9-b56H.js",
    "revision": null
  }, {
    "url": "static/torrentHelpers-DRKpQ7ny.js",
    "revision": null
  }, {
    "url": "static/states-C1wEik_A.js",
    "revision": null
  }, {
    "url": "static/settings-DbAcl2D_.js",
    "revision": null
  }, {
    "url": "static/runtime-DEYqkeYS.js",
    "revision": null
  }, {
    "url": "static/rolldown-runtime-8BhlS34s.js",
    "revision": null
  }, {
    "url": "static/maximize-2-CFLH40jb.js",
    "revision": null
  }, {
    "url": "static/localPrefs-C6BvauQz.js",
    "revision": null
  }, {
    "url": "static/index-DRk8dTQ1.css",
    "revision": null
  }, {
    "url": "static/index-BHJoQ4Vd.js",
    "revision": null
  }, {
    "url": "static/hosts-CMo7Owjk.js",
    "revision": null
  }, {
    "url": "static/hls-BPUeZg91.js",
    "revision": null
  }, {
    "url": "static/heroui-B89MojtZ.js",
    "revision": null
  }, {
    "url": "static/heart-eIbf_rFZ.js",
    "revision": null
  }, {
    "url": "static/gauge-acvAxSwW.js",
    "revision": null
  }, {
    "url": "static/film-AcyTVLhn.js",
    "revision": null
  }, {
    "url": "static/ellipsis-CYZZK6Pr.js",
    "revision": null
  }, {
    "url": "static/createLucideIcon-C43PjgQO.js",
    "revision": null
  }, {
    "url": "static/clapperboard-B-GxjibJ.js",
    "revision": null
  }, {
    "url": "static/circle-alert-DGckt2lj.js",
    "revision": null
  }, {
    "url": "static/authCredentials-DtSl6crY.js",
    "revision": null
  }, {
    "url": "static/VideoPlayer-DjKfBaIm.js",
    "revision": null
  }, {
    "url": "static/UnsafeButton-DVPFr7gx.js",
    "revision": null
  }, {
    "url": "static/SettingsDialog-CbfTFzHr.js",
    "revision": null
  }, {
    "url": "static/ServerStatusDialog-BUSpBoLG.js",
    "revision": null
  }, {
    "url": "static/SearchDialog-C-PLM6nj.js",
    "revision": null
  }, {
    "url": "static/RemoveAllDialog-CMnBM4OD.js",
    "revision": null
  }, {
    "url": "static/PosterPicker-Bb48F0Ej.js",
    "revision": null
  }, {
    "url": "static/PWAInstallationGuide-CJFYQx6w.js",
    "revision": null
  }, {
    "url": "static/MultiAddDialog-ZhBzSYbQ.js",
    "revision": null
  }, {
    "url": "static/ModalOpenContext-DSBubJ8F.js",
    "revision": null
  }, {
    "url": "static/ImportLibraryDialog-DzyaEWxD.js",
    "revision": null
  }, {
    "url": "static/ExportLibraryDialog-6-OHyAny.js",
    "revision": null
  }, {
    "url": "static/EditTorrentDialog-Bs6BGWOQ.js",
    "revision": null
  }, {
    "url": "static/DonateSnackbar-CUbcpbUE.js",
    "revision": null
  }, {
    "url": "static/DonateDialog-DvutGoco.js",
    "revision": null
  }, {
    "url": "static/DetailsDialog-Dsnwirw2.js",
    "revision": null
  }, {
    "url": "static/CommandPalette-DMChVgEA.js",
    "revision": null
  }, {
    "url": "static/CloseServerDialog-m9eu-syy.js",
    "revision": null
  }, {
    "url": "static/CategoriesDrawer-BfMkLitj.js",
    "revision": null
  }, {
    "url": "static/AppDialog-CS7qf9ul.js",
    "revision": null
  }, {
    "url": "static/AndroidInstallBanner-D_0-mrYt.js",
    "revision": null
  }, {
    "url": "static/AddDialog-9JztJmLC.js",
    "revision": null
  }, {
    "url": "static/AboutDialog-D9-8azfW.js",
    "revision": null
  }], {});
  workbox.cleanupOutdatedCaches();
  workbox.registerRoute(new workbox.NavigationRoute(workbox.createHandlerBoundToURL("index.html"), {
    denylist: [/^\/stream/, /^\/torrent/, /^\/torrents/, /^\/cache/, /^\/settings/, /^\/echo/, /^\/gst/, /^\/ffp/, /^\/download/, /^\/viewed/, /^\/search/, /^\/tmdb/, /^\/torznab/, /^\/storage/, /^\/waf/, /^\/mcp/, /^\/shutdown/, /^\/playlistall/, /^\/swagger/, /^\/stat/, /^\/magnets/]
  }));

}));
