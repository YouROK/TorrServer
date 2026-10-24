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
    "revision": "3a32070a4a61b2d9a40b692a2de82c76"
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
    "url": "static/vendor-C6Z8iPso.js",
    "revision": null
  }, {
    "url": "static/useTranslation-D4EcR300.js",
    "revision": null
  }, {
    "url": "static/useTorrentDetail-vA0e06aJ.js",
    "revision": null
  }, {
    "url": "static/torrsLink-CrcAIGZk.js",
    "revision": null
  }, {
    "url": "static/torrents-zWksbVcd.js",
    "revision": null
  }, {
    "url": "static/torrentHelpers-BHhY2_sD.js",
    "revision": null
  }, {
    "url": "static/states-C1wEik_A.js",
    "revision": null
  }, {
    "url": "static/settings-6RhC3ULB.js",
    "revision": null
  }, {
    "url": "static/runtime-T4YatWyc.js",
    "revision": null
  }, {
    "url": "static/rolldown-runtime-CNOWq4Kh.js",
    "revision": null
  }, {
    "url": "static/maximize-2-DqMAUFwG.js",
    "revision": null
  }, {
    "url": "static/localPrefs-C6BvauQz.js",
    "revision": null
  }, {
    "url": "static/index-mh-wt7bV.js",
    "revision": null
  }, {
    "url": "static/index-Ce2_A0_4.css",
    "revision": null
  }, {
    "url": "static/hosts-CEyv4OUe.js",
    "revision": null
  }, {
    "url": "static/hls-D6MMMgcI.js",
    "revision": null
  }, {
    "url": "static/heroui-Cr_NkE6V.js",
    "revision": null
  }, {
    "url": "static/heart-L04XVXcH.js",
    "revision": null
  }, {
    "url": "static/gauge-Bxrn1Gut.js",
    "revision": null
  }, {
    "url": "static/film-DmxuEfjg.js",
    "revision": null
  }, {
    "url": "static/createLucideIcon-B9oi2O-5.js",
    "revision": null
  }, {
    "url": "static/clapperboard-BvSQgFlY.js",
    "revision": null
  }, {
    "url": "static/circle-alert-BS4KXSzd.js",
    "revision": null
  }, {
    "url": "static/authCredentials-D5Z2iq-D.js",
    "revision": null
  }, {
    "url": "static/VideoPlayer-t0oIkjpJ.js",
    "revision": null
  }, {
    "url": "static/UnsafeButton-B5REKvcb.js",
    "revision": null
  }, {
    "url": "static/SettingsDialog-CnTjLd0t.js",
    "revision": null
  }, {
    "url": "static/ServerStatusDialog-5ZrmyPxY.js",
    "revision": null
  }, {
    "url": "static/SearchDialog-Ud0-GmPW.js",
    "revision": null
  }, {
    "url": "static/RemoveAllDialog-hB2mk2RL.js",
    "revision": null
  }, {
    "url": "static/PosterPicker-obEpG_40.js",
    "revision": null
  }, {
    "url": "static/PWAInstallationGuide-BMa9AvmL.js",
    "revision": null
  }, {
    "url": "static/MultiAddDialog-CmSYnh44.js",
    "revision": null
  }, {
    "url": "static/ModalOpenContext-BWJ1W73F.js",
    "revision": null
  }, {
    "url": "static/ImportLibraryDialog-DegaDaQj.js",
    "revision": null
  }, {
    "url": "static/ExportLibraryDialog-C35wLrY7.js",
    "revision": null
  }, {
    "url": "static/EditTorrentDialog-ZlQg1kRF.js",
    "revision": null
  }, {
    "url": "static/DonateSnackbar-B9tF3pUZ.js",
    "revision": null
  }, {
    "url": "static/DonateDialog-Dr_9FW5O.js",
    "revision": null
  }, {
    "url": "static/DetailsDialog-TxLYFJKJ.js",
    "revision": null
  }, {
    "url": "static/CommandPalette-BXz2HCbz.js",
    "revision": null
  }, {
    "url": "static/CloseServerDialog-DchjMk0h.js",
    "revision": null
  }, {
    "url": "static/CategoriesDrawer-CDtJJeWi.js",
    "revision": null
  }, {
    "url": "static/AppDialog-BuU95Awq.js",
    "revision": null
  }, {
    "url": "static/AndroidInstallBanner-JvV6utgQ.js",
    "revision": null
  }, {
    "url": "static/AddDialog-BjxXLeIf.js",
    "revision": null
  }, {
    "url": "static/AboutDialog-C9WxaH14.js",
    "revision": null
  }], {});
  workbox.cleanupOutdatedCaches();
  workbox.registerRoute(new workbox.NavigationRoute(workbox.createHandlerBoundToURL("index.html"), {
    denylist: [/^\/stream/, /^\/torrent/, /^\/torrents/, /^\/cache/, /^\/settings/, /^\/echo/, /^\/gst/, /^\/ffp/, /^\/download/, /^\/viewed/, /^\/search/, /^\/tmdb/, /^\/torznab/, /^\/storage/, /^\/waf/, /^\/mcp/, /^\/shutdown/, /^\/playlistall/, /^\/swagger/, /^\/stat/, /^\/magnets/, /^\/runtime/, /^\/debug/, /^\/ssl/, /^\/mediabase/]
  }));

}));
