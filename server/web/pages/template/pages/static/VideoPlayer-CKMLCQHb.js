import{i as e}from"./rolldown-runtime-CNOWq4Kh.js";import{E as t,M as n,P as r,_ as i,f as a,g as o,k as s,r as c}from"./heroui-Cr_NkE6V.js";import{a as l}from"./vendor-DnQkfXp7.js";import{n as u,r as d}from"./ModalOpenContext-BWJ1W73F.js";import{a as f,l as p,n as m}from"./authCredentials-6FtKqPVT.js";import{t as h}from"./createLucideIcon-B9oi2O-5.js";import{n as ee,t as te}from"./maximize-2-DqMAUFwG.js";import{t as ne}from"./useTranslation-C1MeWBsQ.js";import{$ as re,Bt as ie,Lt as ae,Rt as oe,Vt as se,Y as ce,Zt as le,_ as ue,a as de,b as fe,d as pe,dt as me,g as he,gt as ge,i as _e,mn as ve,s as ye,st as be,un as xe,y as Se}from"./index-jlj0O1ig.js";import{r as Ce,t as we}from"./useTorrentDetail-B4E9Xko0.js";import{t as Te}from"./hls-D6MMMgcI.js";var Ee={name:`minimize-2`,size:24,node:[[`path`,{d:`m14 10 7-7`,key:`oa77jy`}],[`path`,{d:`M20 10h-6V4`,key:`mjg0md`}],[`path`,{d:`m3 21 7-7`,key:`tjx5ai`}],[`path`,{d:`M4 14h6v6`,key:`rmj7iw`}]]};Ee.node;var De=h(Ee),g=e(r(),1),Oe=new Set([`style`,`children`,`ref`,`key`,`suppressContentEditableWarning`,`suppressHydrationWarning`,`dangerouslySetInnerHTML`]),ke={className:`class`,htmlFor:`for`};function Ae(e){return e.toLowerCase()}function je(e){if(typeof e==`boolean`)return e?``:void 0;if(typeof e!=`function`&&!(typeof e==`object`&&e))return e}function _({react:e,tagName:t,elementClass:n,events:r,displayName:i,defaultProps:a,toAttributeName:o=Ae,toAttributeValue:s=je}){let c=Number.parseInt(e.version)>=19,l=e.forwardRef((i,l)=>{let u=e.useRef(null),d=e.useRef(new Map),f={},p={},m={},h={};for(let[e,t]of Object.entries(i)){if(Oe.has(e)){m[e]=t;continue}let r=o(ke[e]??e);if(n.prototype&&e in n.prototype&&!(e in(globalThis.HTMLElement?.prototype??{}))&&!n.observedAttributes?.some(e=>e===r)){h[e]=t;continue}if(e.startsWith(`on`)){f[e]=t;continue}let i=s(t);r&&i!=null&&(p[r]=String(i),c||(m[r]=i)),r&&c&&(m[r]=i===je(t)?t:i)}if(typeof window<`u`){for(let t in f){let n=f[t],i=t.endsWith(`Capture`),a=(r?.[t]??t.slice(2).toLowerCase()).slice(0,i?-7:void 0);e.useLayoutEffect(()=>{let e=u?.current;if(e&&typeof n==`function`)return e.addEventListener(a,n,i),()=>{e.removeEventListener(a,n,i)}},[u?.current,n])}e.useLayoutEffect(()=>{if(u.current===null)return;let e=new Map;for(let t in h)Me(u.current,t,h[t]),d.current.delete(t),e.set(t,h[t]);for(let[e,t]of d.current)Me(u.current,e,void 0);d.current=e})}if(typeof window>`u`&&n?.getTemplateHTML&&n?.shadowRootOptions){let{mode:t,delegatesFocus:r}=n.shadowRootOptions;m.children=[e.createElement(`template`,{shadowrootmode:t,shadowrootdelegatesfocus:r,dangerouslySetInnerHTML:{__html:n.getTemplateHTML(p,i)},key:`ce-la-react-ssr-template-shadow-root`}),m.children]}return e.createElement(t,{...a,...m,ref:e.useCallback(e=>{u.current=e,typeof l==`function`?l(e):l!==null&&(l.current=e)},[l])},m.children)});return l.displayName=i??n.name,l}function Me(e,t,n){e[t]=n,n==null&&t in(globalThis.HTMLElement?.prototype??{})&&e.removeAttribute(t)}var v={MEDIA_PLAY_REQUEST:`mediaplayrequest`,MEDIA_PAUSE_REQUEST:`mediapauserequest`,MEDIA_MUTE_REQUEST:`mediamuterequest`,MEDIA_UNMUTE_REQUEST:`mediaunmuterequest`,MEDIA_LOOP_REQUEST:`medialooprequest`,MEDIA_VOLUME_REQUEST:`mediavolumerequest`,MEDIA_SEEK_REQUEST:`mediaseekrequest`,MEDIA_AIRPLAY_REQUEST:`mediaairplayrequest`,MEDIA_ENTER_FULLSCREEN_REQUEST:`mediaenterfullscreenrequest`,MEDIA_EXIT_FULLSCREEN_REQUEST:`mediaexitfullscreenrequest`,MEDIA_PREVIEW_REQUEST:`mediapreviewrequest`,MEDIA_ENTER_PIP_REQUEST:`mediaenterpiprequest`,MEDIA_EXIT_PIP_REQUEST:`mediaexitpiprequest`,MEDIA_ENTER_CAST_REQUEST:`mediaentercastrequest`,MEDIA_EXIT_CAST_REQUEST:`mediaexitcastrequest`,MEDIA_SHOW_TEXT_TRACKS_REQUEST:`mediashowtexttracksrequest`,MEDIA_HIDE_TEXT_TRACKS_REQUEST:`mediahidetexttracksrequest`,MEDIA_SHOW_SUBTITLES_REQUEST:`mediashowsubtitlesrequest`,MEDIA_DISABLE_SUBTITLES_REQUEST:`mediadisablesubtitlesrequest`,MEDIA_TOGGLE_SUBTITLES_REQUEST:`mediatogglesubtitlesrequest`,MEDIA_PLAYBACK_RATE_REQUEST:`mediaplaybackraterequest`,MEDIA_RENDITION_REQUEST:`mediarenditionrequest`,MEDIA_AUDIO_TRACK_REQUEST:`mediaaudiotrackrequest`,MEDIA_SEEK_TO_LIVE_REQUEST:`mediaseektoliverequest`,REGISTER_MEDIA_STATE_RECEIVER:`registermediastatereceiver`,UNREGISTER_MEDIA_STATE_RECEIVER:`unregistermediastatereceiver`},y={MEDIA_CHROME_ATTRIBUTES:`mediachromeattributes`,MEDIA_CONTROLLER:`mediacontroller`},Ne={MEDIA_AIRPLAY_UNAVAILABLE:`mediaAirplayUnavailable`,MEDIA_AUDIO_TRACK_ENABLED:`mediaAudioTrackEnabled`,MEDIA_AUDIO_TRACK_LIST:`mediaAudioTrackList`,MEDIA_AUDIO_TRACK_UNAVAILABLE:`mediaAudioTrackUnavailable`,MEDIA_BUFFERED:`mediaBuffered`,MEDIA_CAST_UNAVAILABLE:`mediaCastUnavailable`,MEDIA_CHAPTERS_CUES:`mediaChaptersCues`,MEDIA_CURRENT_TIME:`mediaCurrentTime`,MEDIA_DURATION:`mediaDuration`,MEDIA_ENDED:`mediaEnded`,MEDIA_ERROR:`mediaError`,MEDIA_ERROR_CODE:`mediaErrorCode`,MEDIA_ERROR_MESSAGE:`mediaErrorMessage`,MEDIA_FULLSCREEN_UNAVAILABLE:`mediaFullscreenUnavailable`,MEDIA_HAS_PLAYED:`mediaHasPlayed`,MEDIA_HEIGHT:`mediaHeight`,MEDIA_IS_AIRPLAYING:`mediaIsAirplaying`,MEDIA_IS_CASTING:`mediaIsCasting`,MEDIA_IS_FULLSCREEN:`mediaIsFullscreen`,MEDIA_IS_PIP:`mediaIsPip`,MEDIA_LOADING:`mediaLoading`,MEDIA_MUTED:`mediaMuted`,MEDIA_LOOP:`mediaLoop`,MEDIA_PAUSED:`mediaPaused`,MEDIA_PIP_UNAVAILABLE:`mediaPipUnavailable`,MEDIA_PLAYBACK_RATE:`mediaPlaybackRate`,MEDIA_PREVIEW_CHAPTER:`mediaPreviewChapter`,MEDIA_PREVIEW_COORDS:`mediaPreviewCoords`,MEDIA_PREVIEW_IMAGE:`mediaPreviewImage`,MEDIA_PREVIEW_TIME:`mediaPreviewTime`,MEDIA_RENDITION_LIST:`mediaRenditionList`,MEDIA_RENDITION_SELECTED:`mediaRenditionSelected`,MEDIA_RENDITION_UNAVAILABLE:`mediaRenditionUnavailable`,MEDIA_SEEKABLE:`mediaSeekable`,MEDIA_STREAM_TYPE:`mediaStreamType`,MEDIA_SUBTITLES_LIST:`mediaSubtitlesList`,MEDIA_SUBTITLES_SHOWING:`mediaSubtitlesShowing`,MEDIA_TARGET_LIVE_WINDOW:`mediaTargetLiveWindow`,MEDIA_TIME_IS_LIVE:`mediaTimeIsLive`,MEDIA_VOLUME:`mediaVolume`,MEDIA_VOLUME_LEVEL:`mediaVolumeLevel`,MEDIA_VOLUME_UNAVAILABLE:`mediaVolumeUnavailable`,MEDIA_LANG:`mediaLang`,MEDIA_WIDTH:`mediaWidth`},Pe=Object.entries(Ne),b=Pe.reduce((e,[t,n])=>(e[t]=n.toLowerCase(),e),{}),Fe=Pe.reduce((e,[t,n])=>(e[t]=n.toLowerCase(),e),{USER_INACTIVE_CHANGE:`userinactivechange`,BREAKPOINTS_CHANGE:`breakpointchange`,BREAKPOINTS_COMPUTED:`breakpointscomputed`});Object.entries(Fe).reduce((e,[t,n])=>{let r=b[t];return r&&(e[n]=r),e},{userinactivechange:`userinactive`});var x=Object.entries(b).reduce((e,[t,n])=>{let r=Fe[t];return r&&(e[n]=r),e},{userinactive:`userinactivechange`}),S={SUBTITLES:`subtitles`,CAPTIONS:`captions`,DESCRIPTIONS:`descriptions`,CHAPTERS:`chapters`,METADATA:`metadata`},Ie={DISABLED:`disabled`,HIDDEN:`hidden`,SHOWING:`showing`},Le={MOUSE:`mouse`,PEN:`pen`,TOUCH:`touch`},C={UNAVAILABLE:`unavailable`,UNSUPPORTED:`unsupported`},w={LIVE:`live`,ON_DEMAND:`on-demand`,UNKNOWN:`unknown`},T={INLINE:`inline`,FULLSCREEN:`fullscreen`,PICTURE_IN_PICTURE:`picture-in-picture`};function Re(e){return e?.map(ze).join(` `)}function ze(e){if(e){let{id:t,width:n,height:r}=e;return[t,n,r].filter(e=>e!=null).join(`:`)}}function Be(e){return e?.map(Ve).join(` `)}function Ve(e){if(e){let{id:t,kind:n,language:r,label:i}=e;return[t,n,r,i].filter(e=>e!=null).join(`:`)}}function He(e){return typeof e==`number`&&!Number.isNaN(e)&&Number.isFinite(e)}var Ue=e=>new Promise(t=>setTimeout(t,e)),We={"&":`&amp;`,"<":`&lt;`,">":`&gt;`,'"':`&quot;`,"'":`&#39;`};function Ge(e){return String(e).replace(/[&<>"']/g,e=>We[e])}var Ke={en:{"Start airplay":`Start airplay`,"Stop airplay":`Stop airplay`,Audio:`Audio`,Captions:`Captions`,"Enable captions":`Enable captions`,"Disable captions":`Disable captions`,"Start casting":`Start casting`,"Stop casting":`Stop casting`,"Enter fullscreen mode":`Enter fullscreen mode`,"Exit fullscreen mode":`Exit fullscreen mode`,Mute:`Mute`,Unmute:`Unmute`,Loop:`Loop`,"Enter picture in picture mode":`Enter picture in picture mode`,"Exit picture in picture mode":`Exit picture in picture mode`,Play:`Play`,Pause:`Pause`,"Playback rate":`Playback rate`,"Playback rate {playbackRate}":`Playback rate {playbackRate}`,Quality:`Quality`,"Seek backward":`Seek backward`,"Seek forward":`Seek forward`,Settings:`Settings`,Auto:`Auto`,"audio player":`audio player`,"video player":`video player`,volume:`volume`,seek:`seek`,"closed captions":`closed captions`,"current playback rate":`current playback rate`,"playback time":`playback time`,"media loading":`media loading`,settings:`settings`,"audio tracks":`audio tracks`,quality:`quality`,play:`play`,pause:`pause`,mute:`mute`,unmute:`unmute`,"chapter: {chapterName}":`chapter: {chapterName}`,live:`live`,Off:`Off`,"start airplay":`start airplay`,"stop airplay":`stop airplay`,"start casting":`start casting`,"stop casting":`stop casting`,"enter fullscreen mode":`enter fullscreen mode`,"exit fullscreen mode":`exit fullscreen mode`,"enter picture in picture mode":`enter picture in picture mode`,"exit picture in picture mode":`exit picture in picture mode`,"seek to live":`seek to live`,"playing live":`playing live`,"seek back {seekOffset} seconds":`seek back {seekOffset} seconds`,"seek forward {seekOffset} seconds":`seek forward {seekOffset} seconds`,"Network Error":`Network Error`,"Decode Error":`Decode Error`,"Source Not Supported":`Source Not Supported`,"Encryption Error":`Encryption Error`,"A network error caused the media download to fail.":`A network error caused the media download to fail.`,"A media error caused playback to be aborted. The media could be corrupt or your browser does not support this format.":`A media error caused playback to be aborted. The media could be corrupt or your browser does not support this format.`,"An unsupported error occurred. The server or network failed, or your browser does not support this format.":`An unsupported error occurred. The server or network failed, or your browser does not support this format.`,"The media is encrypted and there are no keys to decrypt it.":`The media is encrypted and there are no keys to decrypt it.`,hour:`hour`,hours:`hours`,minute:`minute`,minutes:`minutes`,second:`second`,seconds:`seconds`,"{time} remaining":`{time} remaining`,"{currentTime} of {totalTime}":`{currentTime} of {totalTime}`,"video not loaded, unknown time.":`video not loaded, unknown time.`}},qe=globalThis.navigator?.language||`en`,Je=e=>{qe=e},Ye=e=>{let[t]=qe.split(`-`);return Ke[qe]?.[e]||Ke[t]?.[e]||Ke.en?.[e]||e},Xe=()=>{let[e]=qe.split(`-`);return Ke[qe]?qe:Ke[e]?e:`en`},E=(e,t={})=>Ye(e).replace(/\{(\w+)\}/g,(e,n)=>n in t?String(t[n]):`{${n}}`),Ze=[{singular:`hour`,plural:`hours`},{singular:`minute`,plural:`minutes`},{singular:`second`,plural:`seconds`}],Qe=(e,t)=>`${e} ${E(e===1?Ze[t].singular:Ze[t].plural)}`,$e=e=>{if(!He(e))return``;let t=Math.abs(e),n=t!==e,r=new Date(0,0,0,0,0,t,0),i=[r.getHours(),r.getMinutes(),r.getSeconds()].map((e,t)=>e&&Qe(e,t)).filter(e=>e).join(`, `);return n?E(`{time} remaining`,{time:i}):t===0?Qe(0,2):i};function D(e,t){let n=!1;e<0&&(n=!0,e=0-e),e=e<0?0:e;let r=Math.floor(e%60),i=Math.floor(e/60%60),a=Math.floor(e/3600),o=Math.floor(t/60%60),s=Math.floor(t/3600);return(isNaN(e)||e===1/0)&&(a=i=r=`0`),a=a>0||s>0?a+`:`:``,i=((a||o>=10)&&i<10?`0`+i:i)+`:`,r=r<10?`0`+r:r,(n?`-`:``)+a+i+r}Object.freeze({length:0,start(e){let t=e>>>0;if(t>=this.length)throw new DOMException(`Failed to execute 'start' on 'TimeRanges': The index provided (${t}) is greater than or equal to the maximum bound (${this.length}).`);return 0},end(e){let t=e>>>0;if(t>=this.length)throw new DOMException(`Failed to execute 'end' on 'TimeRanges': The index provided (${t}) is greater than or equal to the maximum bound (${this.length}).`);return 0}});var et=class{addEventListener(){}removeEventListener(){}dispatchEvent(){return!0}},tt=class extends et{},nt=class extends tt{constructor(){super(...arguments),this.role=null}},rt=class{observe(){}unobserve(){}disconnect(){}},it={createElement:function(){return new at.HTMLElement},createElementNS:function(){return new at.HTMLElement},addEventListener(){},removeEventListener(){},dispatchEvent(e){return!1}},at={ResizeObserver:rt,document:it,Node:tt,Element:nt,HTMLElement:class extends nt{constructor(){super(...arguments),this.innerHTML=``}get content(){return new at.DocumentFragment}},DocumentFragment:class extends et{},customElements:{get:function(){},define:function(){},whenDefined:function(){}},localStorage:{getItem(e){return null},setItem(e,t){},removeItem(e){}},CustomEvent:function(){},getComputedStyle:function(){},navigator:{languages:[],get userAgent(){return``}},matchMedia(e){return{matches:!1,media:e}},DOMParser:class{parseFromString(e,t){return{body:{textContent:e}}}}},ot=`global`in globalThis&&(globalThis==null?void 0:globalThis.global)===globalThis||typeof window>`u`||window.customElements===void 0,st=Object.keys(at).every(e=>e in globalThis),O=ot&&!st?at:globalThis,k=ot&&!st?it:globalThis.document,ct=new WeakMap,lt=e=>{let t=ct.get(e);return t||ct.set(e,t=new Set),t},ut=new O.ResizeObserver(e=>{for(let t of e)for(let e of lt(t.target))e(t)});function dt(e,t){lt(e).add(t),ut.observe(e)}function ft(e,t){let n=lt(e);n.delete(t),n.size||ut.unobserve(e)}function pt(e){let t={};for(let n of e)t[n.name]=n.value;return t}function mt(e){return ht(e)??bt(e,`media-controller`)}function ht(e){let{MEDIA_CONTROLLER:t}=y,n=e.getAttribute(t);if(n)return St(e)?.getElementById(n)}var gt=(e,t,n=`.value`)=>{let r=e.querySelector(n);r&&(r.textContent=t)},_t=(e,t)=>{let n=`slot[name="${t}"]`,r=e.shadowRoot.querySelector(n);return r?r.children:[]},vt=(e,t)=>_t(e,t)[0],yt=(e,t)=>!e||!t?!1:e?.contains(t)?!0:yt(e,t.getRootNode().host),bt=(e,t)=>e?e.closest(t)||bt(e.getRootNode().host,t):null;function xt(e=document){let t=e?.activeElement;return t?xt(t.shadowRoot)??t:null}function St(e){let t=(e?.getRootNode)?.call(e);return t instanceof ShadowRoot||t instanceof Document?t:null}function Ct(e,{depth:t=3,checkOpacity:n=!0,checkVisibilityCSS:r=!0}={}){if(e.checkVisibility)return e.checkVisibility({checkOpacity:n,checkVisibilityCSS:r});let i=e;for(;i&&t>0;){let e=getComputedStyle(i);if(n&&e.opacity===`0`||r&&e.visibility===`hidden`||e.display===`none`)return!1;i=i.parentElement,t--}return!0}function wt(e,t,n,r){let i=r.x-n.x,a=r.y-n.y,o=i*i+a*a;if(o===0)return 0;let s=((e-n.x)*i+(t-n.y)*a)/o;return Math.max(0,Math.min(1,s))}function A(e,t){return Tt(e,e=>e===t)||Et(e,t)}function Tt(e,t){let n;for(n of e.querySelectorAll(`style:not([media])`)??[]){let e;try{e=n.sheet?.cssRules}catch{continue}for(let n of e??[])if(t(n.selectorText))return n}}function Et(e,t){let n=e.querySelectorAll(`style:not([media])`)??[],r=n?.[n.length-1];if(!r?.sheet)return console.warn(`Media Chrome: No style sheet found on style tag of`,e),{style:{setProperty:()=>{},removeProperty:()=>``,getPropertyValue:()=>``}};let i=r?.sheet.insertRule(`${t}{}`,r.sheet.cssRules.length);return r.sheet.cssRules?.[i]}function j(e,t,n=NaN){let r=e.getAttribute(t);return r==null?n:+r}function M(e,t,n){let r=+n;n==null||Number.isNaN(r)?e.hasAttribute(t)&&e.removeAttribute(t):j(e,t,void 0)!==r&&e.setAttribute(t,`${r}`)}function N(e,t){return e.hasAttribute(t)}function P(e,t,n){n==null?e.hasAttribute(t)&&e.removeAttribute(t):N(e,t)!=n&&e.toggleAttribute(t,n)}function F(e,t,n=null){return e.getAttribute(t)??n}function I(e,t,n){if(n==null){e.hasAttribute(t)&&e.removeAttribute(t);return}let r=`${n}`;F(e,t,void 0)!==r&&e.setAttribute(t,r)}var Dt=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Ot=(e,t,n)=>(Dt(e,t,`read from private field`),n?n.call(e):t.get(e)),kt=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},At=(e,t,n,r)=>(Dt(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),L;function jt(e){return`
    <style>
      :host {
        display: var(--media-control-display, var(--media-gesture-receiver-display, inline-block));
        box-sizing: border-box;
      }
    </style>
  `}var Mt=class extends O.HTMLElement{constructor(){if(super(),kt(this,L,void 0),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}}static get observedAttributes(){return[y.MEDIA_CONTROLLER,b.MEDIA_PAUSED]}attributeChangedCallback(e,t,n){var r,i,a,o;e===y.MEDIA_CONTROLLER&&(t&&((i=(r=Ot(this,L))?.unassociateElement)==null||i.call(r,this),At(this,L,null)),n&&this.isConnected&&(At(this,L,this.getRootNode()?.getElementById(n)),(o=(a=Ot(this,L))?.associateElement)==null||o.call(a,this)))}connectedCallback(){var e,t;this.tabIndex=-1,this.setAttribute(`aria-hidden`,`true`),At(this,L,Nt(this)),this.getAttribute(y.MEDIA_CONTROLLER)&&((t=(e=Ot(this,L))?.associateElement)==null||t.call(e,this)),Ot(this,L)&&(Ot(this,L).addEventListener(`pointerdown`,this),Ot(this,L).addEventListener(`click`,this),Ot(this,L).hasAttribute(`tabindex`)||(Ot(this,L).tabIndex=0))}disconnectedCallback(){var e,t,n,r;this.getAttribute(y.MEDIA_CONTROLLER)&&((t=(e=Ot(this,L))?.unassociateElement)==null||t.call(e,this)),(n=Ot(this,L))==null||n.removeEventListener(`pointerdown`,this),(r=Ot(this,L))==null||r.removeEventListener(`click`,this),At(this,L,null)}handleEvent(e){let t=e.composedPath()?.[0];if([`video`,`media-controller`].includes(t?.localName)){if(e.type===`pointerdown`)this._pointerType=e.pointerType;else if(e.type===`click`){let{clientX:t,clientY:n}=e,{left:r,top:i,width:a,height:o}=this.getBoundingClientRect(),s=t-r,c=n-i;if(s<0||c<0||s>a||c>o||a===0&&o===0)return;let l=this._pointerType||`mouse`;if(this._pointerType=void 0,l===Le.TOUCH){this.handleTap(e);return}if(l===Le.MOUSE||l===Le.PEN){this.handleMouseClick(e);return}}}}get mediaPaused(){return N(this,b.MEDIA_PAUSED)}set mediaPaused(e){P(this,b.MEDIA_PAUSED,e)}handleTap(e){}handleMouseClick(e){let t=this.mediaPaused?v.MEDIA_PLAY_REQUEST:v.MEDIA_PAUSE_REQUEST;this.dispatchEvent(new O.CustomEvent(t,{composed:!0,bubbles:!0}))}};L=new WeakMap,Mt.shadowRootOptions={mode:`open`},Mt.getTemplateHTML=jt;function Nt(e){let t=e.getAttribute(y.MEDIA_CONTROLLER);return t?e.getRootNode()?.getElementById(t):bt(e,`media-controller`)}O.customElements.get(`media-gesture-receiver`)||O.customElements.define(`media-gesture-receiver`,Mt);var Pt=Mt,Ft=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},R=(e,t,n)=>(Ft(e,t,`read from private field`),n?n.call(e):t.get(e)),z=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},It=(e,t,n,r)=>(Ft(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),Lt=(e,t,n)=>(Ft(e,t,`access private method`),n),Rt,zt,Bt,Vt,Ht,Ut,Wt,Gt,Kt,qt,Jt,Yt,Xt,Zt,Qt,$t,en,tn,nn,rn,B={AUDIO:`audio`,AUTOHIDE:`autohide`,BREAKPOINTS:`breakpoints`,GESTURES_DISABLED:`gesturesdisabled`,KEYBOARD_CONTROL:`keyboardcontrol`,NO_AUTOHIDE:`noautohide`,USER_INACTIVE:`userinactive`,AUTOHIDE_OVER_CONTROLS:`autohideovercontrols`};function an(e){return`
    <style>
      
      :host([${b.MEDIA_IS_FULLSCREEN}]) ::slotted([slot=media]) {
        outline: none;
      }

      :host {
        box-sizing: border-box;
        position: relative;
        display: inline-block;
        line-height: 0;
        background-color: var(--media-background-color, #000);
        overflow: hidden;
      }

      :host(:not([${B.AUDIO}])) [part~=layer]:not([part~=media-layer]) {
        position: absolute;
        top: 0;
        left: 0;
        bottom: 0;
        right: 0;
        display: flex;
        flex-flow: column nowrap;
        align-items: start;
        pointer-events: none;
        background: none;
      }

      slot[name=media] {
        display: var(--media-slot-display, contents);
      }

      
      :host([${B.AUDIO}]) slot[name=media] {
        display: var(--media-slot-display, none);
      }

      
      :host([${B.AUDIO}]) [part~=layer][part~=gesture-layer] {
        height: 0;
        display: block;
      }

      
      :host(:not([${B.AUDIO}])[${B.GESTURES_DISABLED}]) ::slotted([slot=gestures-chrome]),
          :host(:not([${B.AUDIO}])[${B.GESTURES_DISABLED}]) media-gesture-receiver[slot=gestures-chrome] {
        display: none;
      }

      
      ::slotted(:not([slot=media]):not([slot=poster]):not(media-loading-indicator):not([role=dialog]):not([hidden])) {
        pointer-events: auto;
      }

      :host(:not([${B.AUDIO}])) *[part~=layer][part~=centered-layer] {
        align-items: center;
        justify-content: center;
      }

      :host(:not([${B.AUDIO}])) ::slotted(media-gesture-receiver[slot=gestures-chrome]),
      :host(:not([${B.AUDIO}])) media-gesture-receiver[slot=gestures-chrome] {
        align-self: stretch;
        flex-grow: 1;
      }

      slot[name=middle-chrome] {
        display: inline;
        flex-grow: 1;
        pointer-events: none;
        background: none;
      }

      
      ::slotted([slot=media]),
      ::slotted([slot=poster]) {
        width: 100%;
        height: 100%;
      }

      
      :host(:not([${B.AUDIO}])) .spacer {
        flex-grow: 1;
      }

      
      :host(:-webkit-full-screen) {
        
        width: 100% !important;
        height: 100% !important;
      }

      
      ::slotted(:not([slot=media]):not([slot=poster]):not([${B.NO_AUTOHIDE}]):not([hidden]):not([role=dialog])) {
        opacity: 1;
        transition: var(--media-control-transition-in, opacity 0.25s);
      }

      
      :host([${B.USER_INACTIVE}]:not([${b.MEDIA_PAUSED}]):not([${b.MEDIA_IS_AIRPLAYING}]):not([${b.MEDIA_IS_CASTING}]):not([${B.AUDIO}])) ::slotted(:not([slot=media]):not([slot=poster]):not([${B.NO_AUTOHIDE}]):not([role=dialog])) {
        opacity: 0;
        transition: var(--media-control-transition-out, opacity 1s);
      }

      :host([${B.USER_INACTIVE}]:not([${B.NO_AUTOHIDE}]):not([${b.MEDIA_PAUSED}]):not([${b.MEDIA_IS_CASTING}]):not([${B.AUDIO}])) ::slotted([slot=media]) {
        cursor: none;
      }

      :host([${B.USER_INACTIVE}][${B.AUTOHIDE_OVER_CONTROLS}]:not([${B.NO_AUTOHIDE}]):not([${b.MEDIA_PAUSED}]):not([${b.MEDIA_IS_CASTING}]):not([${B.AUDIO}])) * {
        --media-cursor: none;
        cursor: none;
      }


      ::slotted(media-control-bar)  {
        align-self: stretch;
      }

      
      
      :host(:not([${B.AUDIO}])[${b.MEDIA_HAS_PLAYED}]) slot[name=poster],
      :host(:not([${B.AUDIO}])[${b.MEDIA_STREAM_TYPE}="${w.ON_DEMAND}"][${b.MEDIA_CURRENT_TIME}]:not([${b.MEDIA_CURRENT_TIME}="0"])) slot[name=poster] {
        display: none;
      }

      ::slotted([role=dialog]) {
        width: 100%;
        height: 100%;
        align-self: center;
      }

      ::slotted([role=menu]) {
        align-self: end;
      }
    </style>

    <slot name="media" part="layer media-layer"></slot>
    <slot name="poster" part="layer poster-layer"></slot>
    <slot name="gestures-chrome" part="layer gesture-layer">
      <media-gesture-receiver slot="gestures-chrome">
        <template shadowrootmode="${Pt.shadowRootOptions.mode}">
          ${Pt.getTemplateHTML({})}
        </template>
      </media-gesture-receiver>
    </slot>
    <span part="layer vertical-layer">
      <slot name="top-chrome" part="top chrome"></slot>
      <slot name="middle-chrome" part="middle chrome"></slot>
      <slot name="centered-chrome" part="layer centered-layer center centered chrome"></slot>
      
      <slot part="bottom chrome"></slot>
    </span>
    <slot name="dialog" part="layer dialog-layer"></slot>
  `}var on=Object.values(b),sn=`sm:384 md:576 lg:768 xl:960`;function cn(e){ln(e.target,e.contentRect.width)}function ln(e,t){if(!e.isConnected)return;let n=un(e.getAttribute(B.BREAKPOINTS)??sn),r=dn(n,t),i=!1;if(Object.keys(n).forEach(t=>{r.includes(t)?e.hasAttribute(`breakpoint${t}`)||(e.setAttribute(`breakpoint${t}`,``),i=!0):e.hasAttribute(`breakpoint${t}`)&&(e.removeAttribute(`breakpoint${t}`),i=!0)}),i){let t=new CustomEvent(Fe.BREAKPOINTS_CHANGE,{detail:r});e.dispatchEvent(t)}e.breakpointsComputed||(e.breakpointsComputed=!0,e.dispatchEvent(new CustomEvent(Fe.BREAKPOINTS_COMPUTED,{bubbles:!0,composed:!0})))}function un(e){let t=e.split(/\s+/);return Object.fromEntries(t.map(e=>e.split(`:`)))}function dn(e,t){return Object.keys(e).filter(n=>t>=parseInt(e[n]))}var fn=class extends O.HTMLElement{constructor(){if(super(),z(this,Kt),z(this,Jt),z(this,Xt),z(this,Qt),z(this,en),z(this,Rt,void 0),z(this,zt,0),z(this,Bt,null),z(this,Vt,null),z(this,Ht,void 0),this.breakpointsComputed=!1,z(this,Ut,e=>{let t=this.media;for(let n of e){if(n.type!==`childList`)continue;let e=n.removedNodes;for(let r of e){if(r.slot!=`media`||n.target!=this)continue;let e=n.previousSibling&&n.previousSibling.previousElementSibling;if(!e||!t)this.mediaUnsetCallback(r);else{let t=e.slot!==`media`;for(;(e=e.previousSibling)!==null;)e.slot==`media`&&(t=!1);t&&this.mediaUnsetCallback(r)}}if(t)for(let e of n.addedNodes)e===t&&this.handleMediaUpdated(t)}}),z(this,Wt,!1),z(this,Gt,e=>{R(this,Wt)||(setTimeout(()=>{cn(e),It(this,Wt,!1)},0),It(this,Wt,!0))}),z(this,nn,void 0),z(this,rn,()=>{R(this,nn).assignedElements({flatten:!0}).length?this.handleMediaUpdated(this.media):R(this,Bt)&&this.mediaUnsetCallback(R(this,Bt))}),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes),t=this.constructor.getTemplateHTML(e);this.shadowRoot.setHTMLUnsafe?this.shadowRoot.setHTMLUnsafe(t):this.shadowRoot.innerHTML=t}It(this,Rt,new MutationObserver(R(this,Ut)))}static get observedAttributes(){return[B.AUTOHIDE,B.GESTURES_DISABLED].concat(on).filter(e=>![b.MEDIA_RENDITION_LIST,b.MEDIA_AUDIO_TRACK_LIST,b.MEDIA_CHAPTERS_CUES,b.MEDIA_WIDTH,b.MEDIA_HEIGHT,b.MEDIA_ERROR,b.MEDIA_ERROR_MESSAGE].includes(e))}attributeChangedCallback(e,t,n){e.toLowerCase()==B.AUTOHIDE&&(this.autohide=n)}get media(){let e=this.querySelector(`:scope > [slot=media]`);return e?.nodeName==`SLOT`&&(e=e.assignedElements({flatten:!0})[0]),e}async handleMediaUpdated(e){e&&(It(this,Bt,e),e.localName.includes(`-`)&&await O.customElements.whenDefined(e.localName),this.mediaSetCallback(e))}connectedCallback(){var e;R(this,Rt).observe(this,{childList:!0,subtree:!0}),dt(this,R(this,Gt));let t=this.getAttribute(B.AUDIO)==null?E(`video player`):E(`audio player`);this.setAttribute(`role`,`region`),this.setAttribute(`aria-label`,t),this.handleMediaUpdated(this.media),this.setAttribute(B.USER_INACTIVE,``),ln(this,this.getBoundingClientRect().width);let n=this.querySelector(`:scope > slot[slot=media]`);n&&(It(this,nn,n),R(this,nn).addEventListener(`slotchange`,R(this,rn))),this.addEventListener(`pointerdown`,this),this.addEventListener(`pointermove`,this),this.addEventListener(`pointerup`,this),this.addEventListener(`mouseleave`,this),this.addEventListener(`keyup`,this),(e=O.window)==null||e.addEventListener(`mouseup`,this)}disconnectedCallback(){var e;ft(this,R(this,Gt)),clearTimeout(R(this,Vt)),R(this,Rt).disconnect(),this.media&&this.mediaUnsetCallback(this.media),(e=O.window)==null||e.removeEventListener(`mouseup`,this),this.removeEventListener(`pointerdown`,this),this.removeEventListener(`pointermove`,this),this.removeEventListener(`pointerup`,this),this.removeEventListener(`mouseleave`,this),this.removeEventListener(`keyup`,this),R(this,nn)&&(R(this,nn).removeEventListener(`slotchange`,R(this,rn)),It(this,nn,null)),It(this,Wt,!1)}mediaSetCallback(e){}mediaUnsetCallback(e){It(this,Bt,null)}handleEvent(e){switch(e.type){case`pointerdown`:It(this,zt,e.timeStamp);break;case`pointermove`:Lt(this,Kt,qt).call(this,e);break;case`pointerup`:Lt(this,Jt,Yt).call(this,e);break;case`mouseleave`:Lt(this,Xt,Zt).call(this);break;case`mouseup`:this.removeAttribute(B.KEYBOARD_CONTROL);break;case`keyup`:Lt(this,en,tn).call(this),this.setAttribute(B.KEYBOARD_CONTROL,``)}}set autohide(e){let t=Number(e);It(this,Ht,isNaN(t)?0:t)}get autohide(){return(R(this,Ht)===void 0?2:R(this,Ht)).toString()}get breakpoints(){return F(this,B.BREAKPOINTS)}set breakpoints(e){I(this,B.BREAKPOINTS,e)}get audio(){return N(this,B.AUDIO)}set audio(e){P(this,B.AUDIO,e)}get gesturesDisabled(){return N(this,B.GESTURES_DISABLED)}set gesturesDisabled(e){P(this,B.GESTURES_DISABLED,e)}get keyboardControl(){return N(this,B.KEYBOARD_CONTROL)}set keyboardControl(e){P(this,B.KEYBOARD_CONTROL,e)}get noAutohide(){return N(this,B.NO_AUTOHIDE)}set noAutohide(e){P(this,B.NO_AUTOHIDE,e)}get autohideOverControls(){return N(this,B.AUTOHIDE_OVER_CONTROLS)}set autohideOverControls(e){P(this,B.AUTOHIDE_OVER_CONTROLS,e)}get userInteractive(){return N(this,B.USER_INACTIVE)}set userInteractive(e){P(this,B.USER_INACTIVE,e)}};Rt=new WeakMap,zt=new WeakMap,Bt=new WeakMap,Vt=new WeakMap,Ht=new WeakMap,Ut=new WeakMap,Wt=new WeakMap,Gt=new WeakMap,Kt=new WeakSet,qt=function(e){if(e.pointerType!==`mouse`&&e.timeStamp-R(this,zt)<250)return;Lt(this,Qt,$t).call(this),clearTimeout(R(this,Vt));let t=this.hasAttribute(B.AUTOHIDE_OVER_CONTROLS);([this,this.media].includes(e.target)||t)&&Lt(this,en,tn).call(this)},Jt=new WeakSet,Yt=function(e){if(e.pointerType===`touch`){let t=!this.hasAttribute(B.USER_INACTIVE);[this,this.media].includes(e.target)&&t?Lt(this,Xt,Zt).call(this):Lt(this,en,tn).call(this)}else e.composedPath().some(e=>[`media-play-button`,`media-fullscreen-button`].includes(e?.localName))&&Lt(this,en,tn).call(this)},Xt=new WeakSet,Zt=function(){if(R(this,Ht)<0||this.hasAttribute(B.USER_INACTIVE))return;this.setAttribute(B.USER_INACTIVE,``);let e=new O.CustomEvent(Fe.USER_INACTIVE_CHANGE,{composed:!0,bubbles:!0,detail:!0});this.dispatchEvent(e)},Qt=new WeakSet,$t=function(){if(!this.hasAttribute(B.USER_INACTIVE))return;this.removeAttribute(B.USER_INACTIVE);let e=new O.CustomEvent(Fe.USER_INACTIVE_CHANGE,{composed:!0,bubbles:!0,detail:!1});this.dispatchEvent(e)},en=new WeakSet,tn=function(){Lt(this,Qt,$t).call(this),clearTimeout(R(this,Vt));let e=parseInt(this.autohide);e<0||It(this,Vt,setTimeout(()=>{Lt(this,Xt,Zt).call(this)},e*1e3))},nn=new WeakMap,rn=new WeakMap,fn.shadowRootOptions={mode:`open`},fn.getTemplateHTML=an,O.customElements.get(`media-container`)||O.customElements.define(`media-container`,fn);var pn=fn,mn=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},hn=(e,t,n)=>(mn(e,t,`read from private field`),n?n.call(e):t.get(e)),gn=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},_n=(e,t,n,r)=>(mn(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),vn=(e,t,n)=>(mn(e,t,`access private method`),n),yn,bn,xn,Sn,Cn,wn,Tn,En,Dn,On,kn,An,jn,Mn,Nn;function Pn(e){return`
    <style>
      :host {
        font: var(--media-font,
          var(--media-font-weight, normal)
          var(--media-font-size, 14px) /
          var(--media-text-content-height, var(--media-control-height, 24px))
          var(--media-font-family, helvetica neue, segoe ui, roboto, arial, sans-serif));
        color: var(--media-text-color, var(--media-primary-color, rgb(238 238 238)));
        display: var(--media-dialog-display, inline-flex);
        justify-content: center;
        align-items: center;
        
        transition-behavior: allow-discrete;
        visibility: hidden;
        opacity: 0;
        transform: translateY(2px) scale(.99);
        pointer-events: none;
      }

      :host([open]) {
        transition: display .2s, visibility 0s, opacity .2s ease-out, transform .15s ease-out;
        visibility: visible;
        opacity: 1;
        transform: translateY(0) scale(1);
        pointer-events: auto;
      }

      #content {
        display: flex;
        position: relative;
        box-sizing: border-box;
        width: min(320px, 100%);
        word-wrap: break-word;
        max-height: 100%;
        overflow: auto;
        text-align: center;
        line-height: 1.4;
      }
    </style>
    ${this.getSlotTemplateHTML(e)}
  `}function Fn(e){return`
    <slot id="content"></slot>
  `}var In={OPEN:`open`,ANCHOR:`anchor`},Ln=class extends O.HTMLElement{constructor(){super(),gn(this,Sn),gn(this,wn),gn(this,En),gn(this,On),gn(this,An),gn(this,Mn),gn(this,yn,!1),gn(this,bn,null),gn(this,xn,null)}static get observedAttributes(){return[In.OPEN,In.ANCHOR]}get open(){return N(this,In.OPEN)}set open(e){P(this,In.OPEN,e)}handleEvent(e){switch(e.type){case`invoke`:vn(this,On,kn).call(this,e);break;case`focusout`:vn(this,An,jn).call(this,e);break;case`keydown`:vn(this,Mn,Nn).call(this,e)}}connectedCallback(){vn(this,Sn,Cn).call(this),this.role||=`dialog`,this.addEventListener(`invoke`,this),this.addEventListener(`focusout`,this),this.addEventListener(`keydown`,this)}disconnectedCallback(){this.removeEventListener(`invoke`,this),this.removeEventListener(`focusout`,this),this.removeEventListener(`keydown`,this)}attributeChangedCallback(e,t,n){vn(this,Sn,Cn).call(this),e===In.OPEN&&n!==t&&(this.open?vn(this,wn,Tn).call(this):vn(this,En,Dn).call(this))}focus(){_n(this,bn,xt());let e=!this.dispatchEvent(new Event(`focus`,{composed:!0,cancelable:!0})),t=!this.dispatchEvent(new Event(`focusin`,{composed:!0,bubbles:!0,cancelable:!0}));e||t||this.querySelector(`[autofocus], [tabindex]:not([tabindex="-1"]), [role="menu"]`)?.focus()}get keysUsed(){return[`Escape`,`Tab`]}};yn=new WeakMap,bn=new WeakMap,xn=new WeakMap,Sn=new WeakSet,Cn=function(){if(!hn(this,yn)&&(_n(this,yn,!0),!this.shadowRoot)){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e),queueMicrotask(()=>{let{style:e}=A(this.shadowRoot,`:host`);e.setProperty(`transition`,`display .15s, visibility .15s, opacity .15s ease-in, transform .15s ease-in`)})}},wn=new WeakSet,Tn=function(){var e;(e=hn(this,xn))==null||e.setAttribute(`aria-expanded`,`true`),this.dispatchEvent(new Event(`open`,{composed:!0,bubbles:!0})),this.addEventListener(`transitionend`,()=>this.focus(),{once:!0})},En=new WeakSet,Dn=function(){var e;(e=hn(this,xn))==null||e.setAttribute(`aria-expanded`,`false`),this.dispatchEvent(new Event(`close`,{composed:!0,bubbles:!0}))},On=new WeakSet,kn=function(e){_n(this,xn,e.relatedTarget),yt(this,e.relatedTarget)||(this.open=!this.open)},An=new WeakSet,jn=function(e){var t;yt(this,e.relatedTarget)||((t=hn(this,bn))==null||t.focus(),hn(this,xn)&&hn(this,xn)!==e.relatedTarget&&this.open&&(this.open=!1))},Mn=new WeakSet,Nn=function(e){var t,n,r,i,a;let{key:o,ctrlKey:s,altKey:c,metaKey:l}=e;s||c||l||this.keysUsed.includes(o)&&(e.preventDefault(),e.stopPropagation(),o===`Tab`?(e.shiftKey?(n=(t=this.previousElementSibling)?.focus)==null||n.call(t):(i=(r=this.nextElementSibling)?.focus)==null||i.call(r),this.blur()):o===`Escape`&&((a=hn(this,bn))==null||a.focus(),this.open=!1))},Ln.shadowRootOptions={mode:`open`},Ln.getTemplateHTML=Pn,Ln.getSlotTemplateHTML=Fn,O.customElements.get(`media-chrome-dialog`)||O.customElements.define(`media-chrome-dialog`,Ln);var Rn=Ln,zn=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Bn=(e,t,n)=>(zn(e,t,`read from private field`),n?n.call(e):t.get(e)),Vn=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Hn,Un,Wn={HOTKEYS:`hotkeys`},Gn=[{keys:[{key:` `,label:`Space`},{key:`k`}],description:`Toggle Playback`},{keys:[{key:`m`}],description:`Toggle mute`},{keys:[{key:`f`}],description:`Toggle fullscreen`},{keys:[{key:`c`}],description:`Toggle captions or subtitles, if available`},{keys:[{key:`p`}],description:`Toggle Picture in Picture`},{keys:[{key:`ArrowLeft`,label:`←`},{key:`j`}],description:`Seek back 10s`},{keys:[{key:`ArrowRight`,label:`→`},{key:`l`}],description:`Seek forward 10s`},{keys:[{key:`ArrowUp`,label:`↑`}],description:`Turn volume up`},{keys:[{key:`ArrowDown`,label:`↓`}],description:`Turn volume down`},{keys:[{key:`<`,label:`< (SHIFT+,)`}],description:`Decrease playback rate`},{keys:[{key:`>`,label:`> (SHIFT+.)`}],description:`Increase playback rate`},{keys:[{key:`shift+/`,label:`SHIFT+/`}],description:`Show keyboard shortcuts`}];function Kn(e){return e===` `?`nospace`:`no${e.toLowerCase()}`}function qn(e){let t=new Set(e?.split(/\s+/));return Gn.map(({keys:e,description:n})=>({keys:e.filter(({key:e})=>!t.has(Kn(e))),description:n})).filter(({keys:e})=>e.length)}function Jn(e){return`
    <style>
      :host {
        position: fixed;
        top: 0;
        left: 0;
        z-index: 9999;
        background: rgb(20 20 30 / .8);
        backdrop-filter: blur(10px);
      }

      #content {
        display: block;
        width: clamp(400px, 40vw, 700px);
        max-width: 90vw;
        text-align: left;
      }

      h2 {
        margin: 0 0 1.5rem 0;
        font-size: 1.5rem;
        font-weight: 500;
        text-align: center;
      }

      .shortcuts-table {
        width: 100%;
        border-collapse: collapse;
      }

      .shortcuts-table tr {
        border-bottom: 1px solid rgba(255, 255, 255, 0.1);
      }

      .shortcuts-table tr:last-child {
        border-bottom: none;
      }

      .shortcuts-table td {
        padding: 0.75rem 0.5rem;
      }

      .shortcuts-table td:first-child {
        text-align: right;
        padding-right: 1rem;
        width: 40%;
        min-width: 120px;
      }

      .shortcuts-table td:last-child {
        padding-left: 1rem;
      }

      .key {
        display: inline-block;
        background: rgba(255, 255, 255, 0.15);
        border: 1px solid rgba(255, 255, 255, 0.2);
        border-radius: 4px;
        padding: 0.25rem 0.5rem;
        font-family: 'Courier New', monospace;
        font-size: 0.9rem;
        font-weight: 500;
        min-width: 1.5rem;
        text-align: center;
        margin: 0 0.2rem;
      }

      .description {
        color: rgba(255, 255, 255, 0.9);
        font-size: 0.95rem;
      }

      .key-combo {
        display: flex;
        align-items: center;
        justify-content: flex-end;
        gap: 0.3rem;
      }

      .key-separator {
        color: rgba(255, 255, 255, 0.5);
        font-size: 0.9rem;
      }
    </style>
    <slot id="content">
      ${Yn(e[Wn.HOTKEYS])}
    </slot>
  `}function Yn(e){return`
    <h2>Keyboard Shortcuts</h2>
    <table class="shortcuts-table">${qn(e).map(({keys:e,description:t})=>`
        <tr>
          <td>
            <div class="key-combo">${e.map(({key:e,label:t=e},n)=>n>0?`<span class="key-separator">or</span><span class="key">${t}</span>`:`<span class="key">${t}</span>`).join(``)}</div>
          </td>
          <td class="description">${t}</td>
        </tr>
      `).join(``)}</table>
  `}var Xn=class extends Ln{constructor(){super(...arguments),Vn(this,Hn,e=>{if(!this.open)return;let t=this.shadowRoot?.querySelector(`#content`);if(!t)return;let n=e.composedPath(),r=n[0]===this||n.includes(this),i=n.includes(t);r&&!i&&(this.open=!1)}),Vn(this,Un,e=>{if(!this.open)return;let t=e.shiftKey&&(e.key===`/`||e.key===`?`);e.key!==`Escape`&&!t||e.ctrlKey||e.altKey||e.metaKey||(this.open=!1,e.preventDefault(),e.stopPropagation())})}static get observedAttributes(){return[...super.observedAttributes,Wn.HOTKEYS]}get hotkeys(){return F(this,Wn.HOTKEYS)??void 0}set hotkeys(e){I(this,Wn.HOTKEYS,e)}connectedCallback(){super.connectedCallback(),this.open&&(this.addEventListener(`click`,Bn(this,Hn)),document.addEventListener(`keydown`,Bn(this,Un)))}disconnectedCallback(){this.removeEventListener(`click`,Bn(this,Hn)),document.removeEventListener(`keydown`,Bn(this,Un))}attributeChangedCallback(e,t,n){if(super.attributeChangedCallback(e,t,n),e===`open`)this.open?(this.addEventListener(`click`,Bn(this,Hn)),document.addEventListener(`keydown`,Bn(this,Un))):(this.removeEventListener(`click`,Bn(this,Hn)),document.removeEventListener(`keydown`,Bn(this,Un)));else if(e===Wn.HOTKEYS&&n!==t){let e=this.shadowRoot?.querySelector(`#content`);e&&(e.innerHTML=Yn(n))}}};Hn=new WeakMap,Un=new WeakMap,Xn.getSlotTemplateHTML=Jn,O.customElements.get(`media-keyboard-shortcuts-dialog`)||O.customElements.define(`media-keyboard-shortcuts-dialog`,Xn);var Zn=Xn,Qn=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},V=(e,t,n)=>(Qn(e,t,`read from private field`),n?n.call(e):t.get(e)),$n=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},er=(e,t,n,r)=>(Qn(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),tr,nr,rr,ir,ar,or,sr=class{constructor(e,t,{defaultValue:n}={defaultValue:void 0}){$n(this,ar),$n(this,tr,void 0),$n(this,nr,void 0),$n(this,rr,void 0),$n(this,ir,new Set),er(this,tr,e),er(this,nr,t),er(this,rr,new Set(n))}[Symbol.iterator](){return V(this,ar,or).values()}get length(){return V(this,ar,or).size}get value(){return[...V(this,ar,or)].join(` `)??``}set value(e){e!==this.value&&(er(this,ir,new Set),this.add(...e?.split(` `)??[]))}toString(){return this.value}item(e){return[...V(this,ar,or)][e]}values(){return V(this,ar,or).values()}forEach(e,t){V(this,ar,or).forEach(e,t)}add(...e){var t;e.forEach(e=>V(this,ir).add(e)),this.value===``&&!V(this,tr)?.hasAttribute(`${V(this,nr)}`)||(t=V(this,tr))==null||t.setAttribute(`${V(this,nr)}`,`${this.value}`)}remove(...e){var t;e.forEach(e=>V(this,ir).delete(e)),(t=V(this,tr))==null||t.setAttribute(`${V(this,nr)}`,`${this.value}`)}contains(e){return V(this,ar,or).has(e)}toggle(e,t){return t===void 0?this.contains(e)?(this.remove(e),!1):(this.add(e),!0):t?(this.add(e),!0):(this.remove(e),!1)}replace(e,t){return this.remove(e),this.add(t),e===t}};tr=new WeakMap,nr=new WeakMap,rr=new WeakMap,ir=new WeakMap,ar=new WeakSet,or=function(){return V(this,ir).size?V(this,ir):V(this,rr)};var cr=(e=``)=>e.split(/\s+/),lr=(e=``)=>{let[t,n,r]=e.split(`:`),i=r?decodeURIComponent(r):void 0;return{kind:t===`cc`?S.CAPTIONS:S.SUBTITLES,language:n,label:i}},ur=(e=``,t={})=>cr(e).map(e=>{let n=lr(e);return{...t,...n}}),dr=e=>e?Array.isArray(e)?e.map(e=>typeof e==`string`?lr(e):e):typeof e==`string`?ur(e):[e]:[],fr=({kind:e,label:t,language:n}={kind:`subtitles`})=>t?`${e===`captions`?`cc`:`sb`}:${n}:${encodeURIComponent(t)}`:n,pr=(e=[])=>Array.prototype.map.call(e,fr).join(` `),mr=(e,t)=>n=>n[e]===t,hr=e=>{let t=Object.entries(e).map(([e,t])=>mr(e,t));return e=>t.every(t=>t(e))},gr=(e,t=[],n=[])=>{let r=dr(n).map(hr);Array.from(t).filter(e=>r.some(t=>t(e))).forEach(t=>{t.mode=e})},_r=(e,t=()=>!0)=>{if(!e?.textTracks)return[];let n=typeof t==`function`?t:hr(t);return Array.from(e.textTracks).filter(n)},vr=e=>!!e.mediaSubtitlesShowing?.length||e.hasAttribute(b.MEDIA_SUBTITLES_SHOWING),yr=e=>{let{media:t,fullscreenElement:n}=e;try{let e=n&&`requestFullscreen`in n?`requestFullscreen`:n&&`webkitRequestFullScreen`in n?`webkitRequestFullScreen`:void 0;if(e){let t=n[e]?.call(n);if(t instanceof Promise)return t.catch(()=>{})}else t?.webkitEnterFullscreen?t.webkitEnterFullscreen():t?.requestFullscreen&&t.requestFullscreen()}catch(e){console.error(e)}},br=`exitFullscreen`in k?`exitFullscreen`:`webkitExitFullscreen`in k?`webkitExitFullscreen`:`webkitCancelFullScreen`in k?`webkitCancelFullScreen`:void 0,xr=e=>{let{documentElement:t}=e;if(br){let e=(t?.[br])?.call(t);if(e instanceof Promise)return e.catch(()=>{})}},Sr=`fullscreenElement`in k?`fullscreenElement`:`webkitFullscreenElement`in k?`webkitFullscreenElement`:void 0,Cr=e=>{let{documentElement:t,media:n}=e,r=t?.[Sr];return!r&&`webkitDisplayingFullscreen`in n&&`webkitPresentationMode`in n&&n.webkitDisplayingFullscreen&&n.webkitPresentationMode===T.FULLSCREEN?n:r},wr=e=>{let{media:t,documentElement:n,fullscreenElement:r=t}=e;if(!t||!n)return!1;let i=Cr(e);if(!i)return!1;if(i===r||i===t)return!0;if(i.localName.includes(`-`)){let e=i.shadowRoot;if(!(Sr in e))return yt(i,r);for(;e?.[Sr];){if(e[Sr]===r)return!0;e=e[Sr]?.shadowRoot}}return!1},Tr=`fullscreenEnabled`in k?`fullscreenEnabled`:`webkitFullscreenEnabled`in k?`webkitFullscreenEnabled`:void 0,Er=e=>{let{documentElement:t,media:n}=e;return!!t?.[Tr]||n&&`webkitSupportsFullscreen`in n},Dr,Or=()=>{var e;return Dr||(Dr=((e=k)?.createElement)?.call(e,`video`),Dr)},kr=async(e=Or())=>{if(!e)return!1;let t=e.volume;e.volume=t/2+.1;let n=new AbortController,r=await Promise.race([Ar(e,n.signal),jr(e,t)]);return n.abort(),r},Ar=(e,t)=>new Promise(n=>{e.addEventListener(`volumechange`,()=>n(!0),{signal:t})}),jr=async(e,t)=>{for(let n=0;n<10;n++){if(e.volume===t)return!1;await Ue(10)}return e.volume!==t},Mr=/.*Version\/.*Safari\/.*/.test(O.navigator.userAgent),Nr=(e=Or())=>O.matchMedia(`(display-mode: standalone)`).matches&&Mr?!1:typeof e?.requestPictureInPicture==`function`,Pr=(e=Or())=>Er({documentElement:k,media:e}),Fr=Pr(),Ir=Nr(),Lr=!!O.WebKitPlaybackTargetAvailabilityEvent,Rr=!!O.chrome,zr=e=>_r(e.media,e=>[S.SUBTITLES,S.CAPTIONS].includes(e.kind)).sort((e,t)=>e.kind>=t.kind?1:-1),Br=e=>_r(e.media,e=>e.mode===Ie.SHOWING&&[S.SUBTITLES,S.CAPTIONS].includes(e.kind)),Vr=(e,t)=>{let n=zr(e),r=Br(e),i=!!r.length;if(n.length){if(t===!1||i&&t!==!0)gr(Ie.DISABLED,n,r);else if(t===!0||!i&&t!==!1){let t=n[0],{options:i}=e;if(!i?.noSubtitlesLangPref){let e=O.localStorage.getItem(`media-chrome-pref-subtitles-lang`),r=e?[e,...O.navigator.languages]:O.navigator.languages,i=n.filter(e=>r.some(t=>e.language.toLowerCase().startsWith(t.split(`-`)[0]))).sort((e,t)=>r.findIndex(t=>e.language.toLowerCase().startsWith(t.split(`-`)[0]))-r.findIndex(e=>t.language.toLowerCase().startsWith(e.split(`-`)[0])));i[0]&&(t=i[0])}let{language:a,label:o,kind:s}=t;gr(Ie.DISABLED,n,r),gr(Ie.SHOWING,n,[{language:a,label:o,kind:s}])}}},Hr=(e,t)=>e===t?!0:e==null||t==null||typeof e!=typeof t?!1:typeof e==`number`&&Number.isNaN(e)&&Number.isNaN(t)?!0:typeof e==`object`?Array.isArray(e)?Ur(e,t):Object.entries(e).every(([e,n])=>e in t&&Hr(n,t[e])):!1,Ur=(e,t)=>{let n=Array.isArray(e),r=Array.isArray(t);return n===r?n||r?e.length===t.length&&e.every((e,n)=>Hr(e,t[n])):!0:!1},Wr=Object.values(w),Gr,Kr=kr().then(e=>(Gr=e,Gr)),qr=async(...e)=>{await Promise.all(e.filter(e=>e).map(async e=>{if(!(`localName`in e&&e instanceof O.HTMLElement))return;let t=e.localName;if(!t.includes(`-`))return;let n=O.customElements.get(t);n&&e instanceof n||(await O.customElements.whenDefined(t),O.customElements.upgrade(e))}))},Jr=new O.DOMParser,Yr=e=>e&&(Jr.parseFromString(e,`text/html`).body.textContent||e),Xr={mediaError:{get(e,t){let{media:n}=e;if(t?.type!==`playing`)return n?.error},mediaEvents:[`emptied`,`error`,`playing`]},mediaErrorCode:{get(e,t){let{media:n}=e;if(t?.type!==`playing`)return n?.error?.code},mediaEvents:[`emptied`,`error`,`playing`]},mediaErrorMessage:{get(e,t){let{media:n}=e;if(t?.type!==`playing`)return n?.error?.message??``},mediaEvents:[`emptied`,`error`,`playing`]},mediaWidth:{get(e){let{media:t}=e;return t?.videoWidth??0},mediaEvents:[`resize`]},mediaHeight:{get(e){let{media:t}=e;return t?.videoHeight??0},mediaEvents:[`resize`]},mediaPaused:{get(e){let{media:t}=e;return t?.paused??!0},set(e,t){var n;let{media:r}=t;r&&(e?r.pause():(n=r.play())==null||n.catch(()=>{}))},mediaEvents:[`play`,`playing`,`pause`,`emptied`]},mediaHasPlayed:{get(e,t){let{media:n}=e;return n?t?t.type===`playing`:!n.paused:!1},mediaEvents:[`playing`,`emptied`]},mediaEnded:{get(e){let{media:t}=e;return t?.ended??!1},mediaEvents:[`seeked`,`ended`,`emptied`]},mediaPlaybackRate:{get(e){let{media:t}=e;return t?.playbackRate??1},set(e,t){let{media:n}=t;n&&Number.isFinite(+e)&&(n.playbackRate=+e)},mediaEvents:[`ratechange`,`loadstart`]},mediaMuted:{get(e){let{media:t}=e;return t?.muted??!1},set(e,t){let{media:n,options:{noMutedPref:r}={}}=t;if(n){n.muted=e;try{let t=O.localStorage.getItem(`media-chrome-pref-muted`)!==null,i=n.hasAttribute(`muted`);if(r){t&&O.localStorage.removeItem(`media-chrome-pref-muted`);return}if(i&&!t)return;O.localStorage.setItem(`media-chrome-pref-muted`,e?`true`:`false`)}catch(e){console.debug(`Error setting muted pref`,e)}}},mediaEvents:[`volumechange`],stateOwnersUpdateHandlers:[(e,t)=>{let{options:{noMutedPref:n}}=t,{media:r}=t;if(!(!r||r.muted||n))try{let n=O.localStorage.getItem(`media-chrome-pref-muted`)===`true`;Xr.mediaMuted.set(n,t),e(n)}catch(e){console.debug(`Error getting muted pref`,e)}}]},mediaLoop:{get(e){let{media:t}=e;return t?.loop},set(e,t){let{media:n}=t;n&&(n.loop=e)},mediaEvents:[`medialooprequest`]},mediaVolume:{get(e){let{media:t}=e;return t?.volume??1},set(e,t){let{media:n,options:{noVolumePref:r}={}}=t;if(n){try{e==null?O.localStorage.removeItem(`media-chrome-pref-volume`):!n.hasAttribute(`muted`)&&!r&&O.localStorage.setItem(`media-chrome-pref-volume`,e.toString())}catch(e){console.debug(`Error setting volume pref`,e)}Number.isFinite(+e)&&(n.volume=+e)}},mediaEvents:[`volumechange`],stateOwnersUpdateHandlers:[(e,t)=>{let{options:{noVolumePref:n}}=t;if(!n)try{let{media:n}=t;if(!n)return;let r=O.localStorage.getItem(`media-chrome-pref-volume`);if(r==null)return;Xr.mediaVolume.set(+r,t),e(+r)}catch(e){console.debug(`Error getting volume pref`,e)}}]},mediaVolumeLevel:{get(e){let{media:t}=e;return t?.volume===void 0?`high`:t.muted||t.volume===0?`off`:t.volume<.5?`low`:t.volume<.75?`medium`:`high`},mediaEvents:[`volumechange`]},mediaCurrentTime:{get(e){let{media:t}=e;return t?.currentTime??0},set(e,t){let{media:n}=t;n&&He(e)&&(n.currentTime=e)},mediaEvents:[`timeupdate`,`loadedmetadata`,`seeking`]},mediaDuration:{get(e){let{media:t,options:{defaultDuration:n}={}}=e;return n&&(!t||!t.duration||Number.isNaN(t.duration)||!Number.isFinite(t.duration))?n:Number.isFinite(t?.duration)?t.duration:NaN},mediaEvents:[`durationchange`,`loadedmetadata`,`emptied`]},mediaLoading:{get(e){let{media:t}=e;return t?.readyState<3},mediaEvents:[`waiting`,`playing`,`emptied`]},mediaSeekable:{get(e){let{media:t}=e;if(!t?.seekable?.length)return;let n=t.seekable.start(0),r=t.seekable.end(t.seekable.length-1);if(n||r)return[Number(n.toFixed(3)),Number(r.toFixed(3))]},mediaEvents:[`loadedmetadata`,`emptied`,`progress`,`seekablechange`]},mediaBuffered:{get(e){let{media:t}=e,n=t?.buffered??[];return Array.from(n).map((e,t)=>[Number(n.start(t).toFixed(3)),Number(n.end(t).toFixed(3))])},mediaEvents:[`progress`,`emptied`]},mediaStreamType:{get(e){let{media:t,options:{defaultStreamType:n}={}}=e,r=[w.LIVE,w.ON_DEMAND].includes(n)?n:void 0;if(!t)return r;let{streamType:i}=t;if(Wr.includes(i))return i===w.UNKNOWN?r:i;let a=t.duration;return a===1/0?w.LIVE:Number.isFinite(a)?w.ON_DEMAND:r},mediaEvents:[`emptied`,`durationchange`,`loadedmetadata`,`streamtypechange`]},mediaTargetLiveWindow:{get(e){let{media:t}=e;if(!t)return NaN;let{targetLiveWindow:n}=t,r=Xr.mediaStreamType.get(e);return(n==null||Number.isNaN(n))&&r===w.LIVE?0:n},mediaEvents:[`emptied`,`durationchange`,`loadedmetadata`,`streamtypechange`,`targetlivewindowchange`]},mediaTimeIsLive:{get(e){let{media:t,options:{liveEdgeOffset:n=10}={}}=e;if(!t)return!1;if(typeof t.liveEdgeStart==`number`)return!Number.isNaN(t.liveEdgeStart)&&t.currentTime>=t.liveEdgeStart;if(Xr.mediaStreamType.get(e)!==w.LIVE)return!1;let r=t.seekable;if(!r)return!0;if(!r.length)return!1;let i=r.end(r.length-1)-n;return t.currentTime>=i},mediaEvents:[`playing`,`timeupdate`,`progress`,`waiting`,`emptied`]},mediaSubtitlesList:{get(e){return zr(e).map(({kind:e,label:t,language:n})=>({kind:e,label:t,language:n}))},mediaEvents:[`loadstart`],textTracksEvents:[`addtrack`,`removetrack`]},mediaSubtitlesShowing:{get(e){return Br(e).map(({kind:e,label:t,language:n})=>({kind:e,label:t,language:n}))},mediaEvents:[`loadstart`],textTracksEvents:[`addtrack`,`removetrack`,`change`],stateOwnersUpdateHandlers:[(e,t)=>{var n,r;let{media:i,options:a}=t;if(!i)return;let o=e=>{a.defaultSubtitles&&(!e||[S.CAPTIONS,S.SUBTITLES].includes(e?.track?.kind))&&Vr(t,!0)};return i.addEventListener(`loadstart`,o),(n=i.textTracks)==null||n.addEventListener(`addtrack`,o),(r=i.textTracks)==null||r.addEventListener(`removetrack`,o),()=>{var e,t;i.removeEventListener(`loadstart`,o),(e=i.textTracks)==null||e.removeEventListener(`addtrack`,o),(t=i.textTracks)==null||t.removeEventListener(`removetrack`,o)}}]},mediaChaptersCues:{get(e){let{media:t}=e;if(!t)return[];let[n]=_r(t,{kind:S.CHAPTERS});return Array.from(n?.cues??[]).map(({text:e,startTime:t,endTime:n})=>({text:Yr(e),startTime:t,endTime:n}))},mediaEvents:[`loadstart`,`loadedmetadata`],textTracksEvents:[`addtrack`,`removetrack`,`change`],stateOwnersUpdateHandlers:[(e,t)=>{let{media:n}=t;if(!n)return;let r=n.querySelector(`track[kind="chapters"][default][src]`),i=n.shadowRoot?.querySelector(`:is(video,audio) > track[kind="chapters"][default][src]`);return r?.addEventListener(`load`,e),i?.addEventListener(`load`,e),()=>{r?.removeEventListener(`load`,e),i?.removeEventListener(`load`,e)}}]},mediaIsPip:{get(e){let{media:t,documentElement:n}=e;if(!t||!n||!n.pictureInPictureElement)return!1;if(n.pictureInPictureElement===t)return!0;if(n.pictureInPictureElement instanceof HTMLMediaElement)return t.localName?.includes(`-`)?yt(t,n.pictureInPictureElement):!1;if(n.pictureInPictureElement.localName.includes(`-`)){let e=n.pictureInPictureElement.shadowRoot;for(;e?.pictureInPictureElement;){if(e.pictureInPictureElement===t)return!0;e=e.pictureInPictureElement?.shadowRoot}}return!1},set(e,t){let{media:n}=t;if(n){if(e){if(!k.pictureInPictureEnabled){console.warn(`MediaChrome: Picture-in-picture is not enabled`);return}if(!n.requestPictureInPicture){console.warn(`MediaChrome: The current media does not support picture-in-picture`);return}let e=()=>{console.warn(`MediaChrome: The media is not ready for picture-in-picture. It must have a readyState > 0.`)};n.requestPictureInPicture().catch(t=>{if(t.code===11){if(!n.src){console.warn(`MediaChrome: The media is not ready for picture-in-picture. It must have a src set.`);return}if(n.readyState===0&&n.preload===`none`){let t=()=>{n.removeEventListener(`loadedmetadata`,r),n.preload=`none`},r=()=>{n.requestPictureInPicture().catch(e),t()};n.addEventListener(`loadedmetadata`,r),n.preload=`metadata`,setTimeout(()=>{n.readyState===0&&e(),t()},1e3)}else throw t}else throw t})}else k.pictureInPictureElement&&k.exitPictureInPicture()}},mediaEvents:[`enterpictureinpicture`,`leavepictureinpicture`]},mediaRenditionList:{get(e){let{media:t}=e;return[...t?.videoRenditions??[]].map(e=>({...e}))},mediaEvents:[`emptied`,`loadstart`],videoRenditionsEvents:[`addrendition`,`removerendition`]},mediaRenditionSelected:{get(e){let{media:t}=e;return t?.videoRenditions?.[t.videoRenditions?.selectedIndex]?.id},set(e,t){let{media:n}=t;if(!n?.videoRenditions){console.warn(`MediaController: Rendition selection not supported by this media.`);return}let r=e,i=Array.prototype.findIndex.call(n.videoRenditions,e=>e.id==r);n.videoRenditions.selectedIndex!=i&&(n.videoRenditions.selectedIndex=i)},mediaEvents:[`emptied`],videoRenditionsEvents:[`addrendition`,`removerendition`,`change`]},mediaAudioTrackList:{get(e){let{media:t}=e;return[...t?.audioTracks??[]]},mediaEvents:[`emptied`,`loadstart`],audioTracksEvents:[`addtrack`,`removetrack`]},mediaAudioTrackEnabled:{get(e){let{media:t}=e;return[...t?.audioTracks??[]].find(e=>e.enabled)?.id},set(e,t){let{media:n}=t;if(!n?.audioTracks){console.warn(`MediaChrome: Audio track selection not supported by this media.`);return}let r=e;for(let e of n.audioTracks)e.enabled=r==e.id},mediaEvents:[`emptied`],audioTracksEvents:[`addtrack`,`removetrack`,`change`]},mediaIsFullscreen:{get(e){return wr(e)},set(e,t,n){var r;e?(yr(t),n.detail&&!t.media?.inert&&((r=t.media)==null||r.focus())):xr(t)},rootEvents:[`fullscreenchange`,`webkitfullscreenchange`],mediaEvents:[`webkitbeginfullscreen`,`webkitendfullscreen`,`webkitpresentationmodechanged`]},mediaIsCasting:{get(e){let{media:t}=e;return!t?.remote||t.remote?.state===`disconnected`?!1:t.remote.state===`connected`},set(e,t){let{media:n}=t;n&&(e&&n.remote?.state!==`disconnected`||(e||n.remote?.state===`connected`)&&(typeof n.remote.prompt==`function`?n.remote.prompt().catch(()=>{}):console.warn(`MediaChrome: Casting is not supported in this environment`)))},remoteEvents:[`connect`,`connecting`,`disconnect`]},mediaIsAirplaying:{get(){return!1},set(e,t){let{media:n}=t;n&&(n.webkitShowPlaybackTargetPicker&&O.WebKitPlaybackTargetAvailabilityEvent?n.webkitShowPlaybackTargetPicker():console.error(`MediaChrome: received a request to select AirPlay but AirPlay is not supported in this environment`))},mediaEvents:[`webkitcurrentplaybacktargetiswirelesschanged`]},mediaFullscreenUnavailable:{get(e){let{media:t}=e;if(!Fr||!Pr(t))return C.UNSUPPORTED}},mediaPipUnavailable:{get(e){let{media:t}=e;if(!Ir||!Nr(t))return C.UNSUPPORTED;if(t?.disablePictureInPicture)return C.UNAVAILABLE}},mediaVolumeUnavailable:{get(e){let{media:t}=e;if(Gr===!1||t?.volume==null)return C.UNSUPPORTED},stateOwnersUpdateHandlers:[e=>{Gr??Kr.then(t=>e(t?void 0:C.UNSUPPORTED))}]},mediaCastUnavailable:{get(e,{availability:t=`not-available`}={}){let{media:n}=e;if(!Rr||!n?.remote?.state)return C.UNSUPPORTED;if(t!=null&&t!==`available`)return C.UNAVAILABLE},stateOwnersUpdateHandlers:[(e,t)=>{var n;let{media:r}=t;if(r)return r.disableRemotePlayback||r.hasAttribute(`disableremoteplayback`)||(n=r?.remote)==null||n.watchAvailability(t=>{e({availability:t?`available`:`not-available`})}).catch(t=>{t.name===`NotSupportedError`?e({availability:null}):e({availability:`not-available`})}),()=>{var e;(e=r?.remote)==null||e.cancelWatchAvailability().catch(()=>{})}}]},mediaAirplayUnavailable:{get(e,t){if(!Lr)return C.UNSUPPORTED;if(t?.availability===`not-available`)return C.UNAVAILABLE},mediaEvents:[`webkitplaybacktargetavailabilitychanged`],stateOwnersUpdateHandlers:[(e,t)=>{var n;let{media:r}=t;if(r)return r.disableRemotePlayback||r.hasAttribute(`disableremoteplayback`)||(n=r?.remote)==null||n.watchAvailability(t=>{e({availability:t?`available`:`not-available`})}).catch(t=>{t.name===`NotSupportedError`?e({availability:null}):e({availability:`not-available`})}),()=>{var e;(e=r?.remote)==null||e.cancelWatchAvailability().catch(()=>{})}}]},mediaRenditionUnavailable:{get(e){let{media:t}=e;if(!t?.videoRenditions)return C.UNSUPPORTED;if(!t.videoRenditions?.length)return C.UNAVAILABLE},mediaEvents:[`emptied`,`loadstart`],videoRenditionsEvents:[`addrendition`,`removerendition`]},mediaAudioTrackUnavailable:{get(e){let{media:t}=e;if(!t?.audioTracks)return C.UNSUPPORTED;if((t.audioTracks?.length??0)<=1)return C.UNAVAILABLE},mediaEvents:[`emptied`,`loadstart`],audioTracksEvents:[`addtrack`,`removetrack`]},mediaLang:{get(e){let{options:{mediaLang:t}={}}=e;return t??`en`}}},Zr={[v.MEDIA_PREVIEW_REQUEST](e,t,{detail:n}){let{media:r}=t,i=n??void 0,a,o;if(r&&i!=null){let[e]=_r(r,{kind:S.METADATA,label:`thumbnails`}),t=Array.prototype.find.call(e?.cues??[],(e,t,n)=>t===0?e.endTime>i:t===n.length-1?e.startTime<=i:e.startTime<=i&&e.endTime>i);if(t){let e=/'^(?:[a-z]+:)?\/\//i.test(t.text)?void 0:r?.querySelector(`track[label="thumbnails"]`)?.src,n=new URL(t.text,e);o=new URLSearchParams(n.hash).get(`#xywh`).split(`,`).map(e=>+e),a=n.href}}let s=e.mediaDuration.get(t),c=e.mediaChaptersCues.get(t).find((e,t,n)=>t===n.length-1&&s===e.endTime?e.startTime<=i&&e.endTime>=i:e.startTime<=i&&e.endTime>i)?.text;return n!=null&&c==null&&(c=``),{mediaPreviewTime:i,mediaPreviewImage:a,mediaPreviewCoords:o,mediaPreviewChapter:c}},[v.MEDIA_PAUSE_REQUEST](e,t){e.mediaPaused.set(!0,t)},[v.MEDIA_PLAY_REQUEST](e,t){let n=e.mediaStreamType.get(t)===w.LIVE,r=!t.options?.noAutoSeekToLive,i=e.mediaTargetLiveWindow.get(t)>0;if(n&&r&&!i){let n=e.mediaSeekable.get(t)?.[1];if(n){let r=n-(t.options?.seekToLiveOffset??0);e.mediaCurrentTime.set(r,t)}}e.mediaPaused.set(!1,t)},[v.MEDIA_PLAYBACK_RATE_REQUEST](e,t,{detail:n}){let r=n;e.mediaPlaybackRate.set(r,t)},[v.MEDIA_MUTE_REQUEST](e,t){e.mediaMuted.set(!0,t)},[v.MEDIA_UNMUTE_REQUEST](e,t){e.mediaVolume.get(t)||e.mediaVolume.set(.25,t),e.mediaMuted.set(!1,t)},[v.MEDIA_LOOP_REQUEST](e,t,{detail:n}){let r=!!n;return e.mediaLoop.set(r,t),{mediaLoop:r}},[v.MEDIA_VOLUME_REQUEST](e,t,{detail:n}){let r=n;r&&e.mediaMuted.get(t)&&e.mediaMuted.set(!1,t),e.mediaVolume.set(r,t)},[v.MEDIA_SEEK_REQUEST](e,t,{detail:n}){let r=n;e.mediaCurrentTime.set(r,t)},[v.MEDIA_SEEK_TO_LIVE_REQUEST](e,t){let n=e.mediaSeekable.get(t)?.[1];if(Number.isNaN(Number(n)))return;let r=n-(t.options?.seekToLiveOffset??0);e.mediaCurrentTime.set(r,t)},[v.MEDIA_SHOW_SUBTITLES_REQUEST](e,t,{detail:n}){let{options:r}=t,i=zr(t),a=dr(n),o=a[0]?.language;o&&!r.noSubtitlesLangPref&&O.localStorage.setItem(`media-chrome-pref-subtitles-lang`,o),gr(Ie.SHOWING,i,a)},[v.MEDIA_DISABLE_SUBTITLES_REQUEST](e,t,{detail:n}){let r=zr(t),i=n??[];gr(Ie.DISABLED,r,i)},[v.MEDIA_TOGGLE_SUBTITLES_REQUEST](e,t,{detail:n}){Vr(t,n)},[v.MEDIA_RENDITION_REQUEST](e,t,{detail:n}){let r=n;e.mediaRenditionSelected.set(r,t)},[v.MEDIA_AUDIO_TRACK_REQUEST](e,t,{detail:n}){let r=n;e.mediaAudioTrackEnabled.set(r,t)},[v.MEDIA_ENTER_PIP_REQUEST](e,t){e.mediaIsFullscreen.get(t)&&e.mediaIsFullscreen.set(!1,t),e.mediaIsPip.set(!0,t)},[v.MEDIA_EXIT_PIP_REQUEST](e,t){e.mediaIsPip.set(!1,t)},[v.MEDIA_ENTER_FULLSCREEN_REQUEST](e,t,n){e.mediaIsPip.get(t)&&e.mediaIsPip.set(!1,t),e.mediaIsFullscreen.set(!0,t,n)},[v.MEDIA_EXIT_FULLSCREEN_REQUEST](e,t){e.mediaIsFullscreen.set(!1,t)},[v.MEDIA_ENTER_CAST_REQUEST](e,t){e.mediaIsFullscreen.get(t)&&e.mediaIsFullscreen.set(!1,t),e.mediaIsCasting.set(!0,t)},[v.MEDIA_EXIT_CAST_REQUEST](e,t){e.mediaIsCasting.set(!1,t)},[v.MEDIA_AIRPLAY_REQUEST](e,t){e.mediaIsAirplaying.set(!0,t)}},Qr=({media:e,fullscreenElement:t,documentElement:n,stateMediator:r=Xr,requestMap:i=Zr,options:a={},monitorStateOwnersOnlyWithSubscriptions:o=!0})=>{let s=[],c={options:{...a}},l=Object.freeze({mediaPreviewTime:void 0,mediaPreviewImage:void 0,mediaPreviewCoords:void 0,mediaPreviewChapter:void 0}),u=e=>{e!=null&&(Hr(e,l)||(l=Object.freeze({...l,...e}),s.forEach(e=>e(l))))},d=()=>{let e=Object.entries(r).reduce((e,[t,{get:n}])=>(e[t]=n(c),e),{});u(e)},f={},p,m=async(e,t)=>{let n=!!p;if(p={...c,...p??{},...e},n)return;await qr(...Object.values(e));let i=s.length>0&&t===0&&o,a=c.media!==p.media,l=c.media?.textTracks!==p.media?.textTracks,m=c.media?.videoRenditions!==p.media?.videoRenditions,h=c.media?.audioTracks!==p.media?.audioTracks,ee=c.media?.remote!==p.media?.remote,te=c.documentElement!==p.documentElement,ne=!!c.media&&(a||i),re=!!c.media?.textTracks&&(l||i),ie=!!c.media?.videoRenditions&&(m||i),ae=!!c.media?.audioTracks&&(h||i),oe=!!c.media?.remote&&(ee||i),se=!!c.documentElement&&(te||i),ce=ne||re||ie||ae||oe||se,le=s.length===0&&t===1&&o,ue=!!p.media&&(a||le),de=!!p.media?.textTracks&&(l||le),fe=!!p.media?.videoRenditions&&(m||le),pe=!!p.media?.audioTracks&&(h||le),me=!!p.media?.remote&&(ee||le),he=!!p.documentElement&&(te||le),ge=ue||de||fe||pe||me||he;ce||ge?(Object.entries(r).forEach(([e,{get:t,mediaEvents:n=[],textTracksEvents:r=[],videoRenditionsEvents:i=[],audioTracksEvents:a=[],remoteEvents:o=[],rootEvents:s=[],stateOwnersUpdateHandlers:l=[]}])=>{f[e]||(f[e]={});let d=n=>{let r=t(c,n);u({[e]:r})},m;m=f[e].mediaEvents,n.forEach(t=>{m&&ne&&(c.media.removeEventListener(t,m),f[e].mediaEvents=void 0),ue&&(p.media.addEventListener(t,d),f[e].mediaEvents=d)}),m=f[e].textTracksEvents,r.forEach(t=>{var n,r;m&&re&&((n=c.media.textTracks)==null||n.removeEventListener(t,m),f[e].textTracksEvents=void 0),de&&((r=p.media.textTracks)==null||r.addEventListener(t,d),f[e].textTracksEvents=d)}),m=f[e].videoRenditionsEvents,i.forEach(t=>{var n,r;m&&ie&&((n=c.media.videoRenditions)==null||n.removeEventListener(t,m),f[e].videoRenditionsEvents=void 0),fe&&((r=p.media.videoRenditions)==null||r.addEventListener(t,d),f[e].videoRenditionsEvents=d)}),m=f[e].audioTracksEvents,a.forEach(t=>{var n,r;m&&ae&&((n=c.media.audioTracks)==null||n.removeEventListener(t,m),f[e].audioTracksEvents=void 0),pe&&((r=p.media.audioTracks)==null||r.addEventListener(t,d),f[e].audioTracksEvents=d)}),m=f[e].remoteEvents,o.forEach(t=>{var n,r;m&&oe&&((n=c.media.remote)==null||n.removeEventListener(t,m),f[e].remoteEvents=void 0),me&&((r=p.media.remote)==null||r.addEventListener(t,d),f[e].remoteEvents=d)}),m=f[e].rootEvents,s.forEach(t=>{m&&se&&(c.documentElement.removeEventListener(t,m),f[e].rootEvents=void 0),he&&(p.documentElement.addEventListener(t,d),f[e].rootEvents=d)});let h=f[e].stateOwnersUpdateHandlers;if(h&&ce&&(Array.isArray(h)?h:[h]).forEach(e=>{typeof e==`function`&&e()}),ge){let t=l.map(e=>e(d,p)).filter(e=>typeof e==`function`);f[e].stateOwnersUpdateHandlers=t.length===1?t[0]:t}else ce&&(f[e].stateOwnersUpdateHandlers=void 0)}),Object.entries(p).forEach(([e,t])=>{c[e]=t}),d(),p=void 0):(Object.entries(p).forEach(([e,t])=>{c[e]=t}),d(),p=void 0)};return m({media:e,fullscreenElement:t,documentElement:n,options:a}),{dispatch(e){let{type:t,detail:n}=e;i[t]&&l.mediaErrorCode==null?u(i[t](r,c,e)):t===`mediaelementchangerequest`?m({media:n}):t===`fullscreenelementchangerequest`?m({fullscreenElement:n}):t===`documentelementchangerequest`?m({documentElement:n}):t===`optionschangerequest`&&(Object.entries(n??{}).forEach(([e,t])=>{c.options[e]=t}),d())},getState(){return l},subscribe(e){return m({},s.length+1),s.push(e),e(l),()=>{let t=s.indexOf(e);t>=0&&(m({},s.length-1),s.splice(t,1))}}}},$r=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},H=(e,t,n)=>($r(e,t,`read from private field`),n?n.call(e):t.get(e)),ei=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},ti=(e,t,n,r)=>($r(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),ni=(e,t,n)=>($r(e,t,`access private method`),n),ri,ii,U,ai,oi,si,ci,li,ui,di,fi,pi,mi,hi,gi,_i,vi,yi=[`ArrowLeft`,`ArrowRight`,`ArrowUp`,`ArrowDown`,`Enter`,` `,`f`,`m`,`k`,`c`,`l`,`j`,`>`,`<`,`p`],bi=10,xi=.025,Si=.25,Ci=.25,wi=2,W={DEFAULT_SUBTITLES:`defaultsubtitles`,DEFAULT_STREAM_TYPE:`defaultstreamtype`,DEFAULT_DURATION:`defaultduration`,FULLSCREEN_ELEMENT:`fullscreenelement`,HOTKEYS:`hotkeys`,KEYBOARD_BACKWARD_SEEK_OFFSET:`keyboardbackwardseekoffset`,KEYBOARD_FORWARD_SEEK_OFFSET:`keyboardforwardseekoffset`,KEYBOARD_DOWN_VOLUME_STEP:`keyboarddownvolumestep`,KEYBOARD_UP_VOLUME_STEP:`keyboardupvolumestep`,KEYS_USED:`keysused`,LANG:`lang`,LOOP:`loop`,LIVE_EDGE_OFFSET:`liveedgeoffset`,NO_AUTO_SEEK_TO_LIVE:`noautoseektolive`,NO_DEFAULT_STORE:`nodefaultstore`,NO_HOTKEYS:`nohotkeys`,NO_MUTED_PREF:`nomutedpref`,NO_SUBTITLES_LANG_PREF:`nosubtitleslangpref`,NO_VOLUME_PREF:`novolumepref`,SEEK_TO_LIVE_OFFSET:`seektoliveoffset`},Ti=class extends fn{constructor(){super(),ei(this,ui),ei(this,pi),ei(this,hi),ei(this,_i),this.mediaStateReceivers=[],this.associatedElementSubscriptions=new Map,ei(this,ri,new sr(this,W.HOTKEYS)),ei(this,ii,void 0),ei(this,U,void 0),ei(this,ai,null),ei(this,oi,void 0),ei(this,si,void 0),ei(this,ci,e=>{var t;(t=H(this,U))==null||t.dispatch(e)}),ei(this,li,void 0),ei(this,fi,e=>{let{key:t,shiftKey:n}=e;n&&(t===`/`||t===`?`)||yi.includes(t)?this.keyboardShortcutHandler(e):this.removeEventListener(`keyup`,H(this,fi))}),this.associateElement(this);let e={};ti(this,oi,t=>{Object.entries(t).forEach(([t,n])=>{if(t in e&&e[t]===n)return;this.propagateMediaState(t,n);let r=t.toLowerCase(),i=new O.CustomEvent(x[r],{composed:!0,detail:n});this.dispatchEvent(i)}),e=t})}static get observedAttributes(){return super.observedAttributes.concat(W.NO_HOTKEYS,W.HOTKEYS,W.DEFAULT_STREAM_TYPE,W.DEFAULT_SUBTITLES,W.DEFAULT_DURATION,W.NO_MUTED_PREF,W.NO_VOLUME_PREF,W.LANG,W.LOOP,W.LIVE_EDGE_OFFSET,W.SEEK_TO_LIVE_OFFSET,W.NO_AUTO_SEEK_TO_LIVE)}get mediaStore(){return H(this,U)}set mediaStore(e){var t;H(this,U)&&((t=H(this,si))==null||t.call(this),ti(this,si,void 0)),ti(this,U,e),!H(this,U)&&!this.hasAttribute(W.NO_DEFAULT_STORE)?ni(this,ui,di).call(this):ti(this,si,H(this,U)?.subscribe(H(this,oi)))}get fullscreenElement(){return H(this,ii)??this}set fullscreenElement(e){var t;this.hasAttribute(W.FULLSCREEN_ELEMENT)&&this.removeAttribute(W.FULLSCREEN_ELEMENT),ti(this,ii,e),(t=H(this,U))==null||t.dispatch({type:`fullscreenelementchangerequest`,detail:this.fullscreenElement})}get defaultSubtitles(){return N(this,W.DEFAULT_SUBTITLES)}set defaultSubtitles(e){P(this,W.DEFAULT_SUBTITLES,e)}get defaultStreamType(){return F(this,W.DEFAULT_STREAM_TYPE)}set defaultStreamType(e){I(this,W.DEFAULT_STREAM_TYPE,e)}get defaultDuration(){return j(this,W.DEFAULT_DURATION)}set defaultDuration(e){M(this,W.DEFAULT_DURATION,e)}get noHotkeys(){return N(this,W.NO_HOTKEYS)}set noHotkeys(e){P(this,W.NO_HOTKEYS,e)}get keysUsed(){return F(this,W.KEYS_USED)}set keysUsed(e){I(this,W.KEYS_USED,e)}get liveEdgeOffset(){return j(this,W.LIVE_EDGE_OFFSET)}set liveEdgeOffset(e){M(this,W.LIVE_EDGE_OFFSET,e)}get noAutoSeekToLive(){return N(this,W.NO_AUTO_SEEK_TO_LIVE)}set noAutoSeekToLive(e){P(this,W.NO_AUTO_SEEK_TO_LIVE,e)}get noVolumePref(){return N(this,W.NO_VOLUME_PREF)}set noVolumePref(e){P(this,W.NO_VOLUME_PREF,e)}get noMutedPref(){return N(this,W.NO_MUTED_PREF)}set noMutedPref(e){P(this,W.NO_MUTED_PREF,e)}get noSubtitlesLangPref(){return N(this,W.NO_SUBTITLES_LANG_PREF)}set noSubtitlesLangPref(e){P(this,W.NO_SUBTITLES_LANG_PREF,e)}get noDefaultStore(){return N(this,W.NO_DEFAULT_STORE)}set noDefaultStore(e){P(this,W.NO_DEFAULT_STORE,e)}get resolvedLang(){return Xe()}attributeChangedCallback(e,t,n){var r,i,a,o,s,c,l,u,d,f;if(super.attributeChangedCallback(e,t,n),e===W.NO_HOTKEYS)n!==t&&n===``?(this.hasAttribute(W.HOTKEYS)&&console.warn("Media Chrome: Both `hotkeys` and `nohotkeys` have been set. All hotkeys will be disabled."),this.disableHotkeys()):n!==t&&n===null&&this.enableHotkeys(),ni(this,hi,gi).call(this);else if(e===W.HOTKEYS)H(this,ri).value=n,ni(this,hi,gi).call(this);else if(e===W.DEFAULT_SUBTITLES&&n!==t)(r=H(this,U))==null||r.dispatch({type:`optionschangerequest`,detail:{defaultSubtitles:this.hasAttribute(W.DEFAULT_SUBTITLES)}});else if(e===W.DEFAULT_STREAM_TYPE)(i=H(this,U))==null||i.dispatch({type:`optionschangerequest`,detail:{defaultStreamType:this.getAttribute(W.DEFAULT_STREAM_TYPE)??void 0}});else if(e===W.LIVE_EDGE_OFFSET&&n!==t)(a=H(this,U))==null||a.dispatch({type:`optionschangerequest`,detail:{liveEdgeOffset:this.hasAttribute(W.LIVE_EDGE_OFFSET)?+this.getAttribute(W.LIVE_EDGE_OFFSET):void 0,seekToLiveOffset:this.hasAttribute(W.SEEK_TO_LIVE_OFFSET)?+this.getAttribute(W.SEEK_TO_LIVE_OFFSET):this.hasAttribute(W.LIVE_EDGE_OFFSET)?+this.getAttribute(W.LIVE_EDGE_OFFSET):void 0}});else if(e===W.SEEK_TO_LIVE_OFFSET&&n!==t)(o=H(this,U))==null||o.dispatch({type:`optionschangerequest`,detail:{seekToLiveOffset:this.hasAttribute(W.SEEK_TO_LIVE_OFFSET)?+this.getAttribute(W.SEEK_TO_LIVE_OFFSET):this.hasAttribute(W.LIVE_EDGE_OFFSET)?+this.getAttribute(W.LIVE_EDGE_OFFSET):void 0}});else if(e===W.NO_AUTO_SEEK_TO_LIVE)(s=H(this,U))==null||s.dispatch({type:`optionschangerequest`,detail:{noAutoSeekToLive:this.hasAttribute(W.NO_AUTO_SEEK_TO_LIVE)}});else if(e===W.FULLSCREEN_ELEMENT){let e=n?this.getRootNode()?.getElementById(n):void 0;ti(this,ii,e),(c=H(this,U))==null||c.dispatch({type:`fullscreenelementchangerequest`,detail:this.fullscreenElement})}else e===W.LANG&&n!==t?(Je(n),(l=H(this,U))==null||l.dispatch({type:`optionschangerequest`,detail:{mediaLang:n}})):e===W.LOOP&&n!==t?(u=H(this,U))==null||u.dispatch({type:v.MEDIA_LOOP_REQUEST,detail:n!=null}):e===W.NO_VOLUME_PREF&&n!==t?(d=H(this,U))==null||d.dispatch({type:`optionschangerequest`,detail:{noVolumePref:this.hasAttribute(W.NO_VOLUME_PREF)}}):e===W.NO_MUTED_PREF&&n!==t&&((f=H(this,U))==null||f.dispatch({type:`optionschangerequest`,detail:{noMutedPref:this.hasAttribute(W.NO_MUTED_PREF)}}))}connectedCallback(){var e,t;this.associateElement(this),!H(this,U)&&!this.hasAttribute(W.NO_DEFAULT_STORE)&&ni(this,ui,di).call(this),(e=H(this,U))==null||e.dispatch({type:`documentelementchangerequest`,detail:k}),(t=H(this,U))==null||t.dispatch({type:`fullscreenelementchangerequest`,detail:this.fullscreenElement}),super.connectedCallback(),H(this,U)&&!H(this,si)&&ti(this,si,H(this,U)?.subscribe(H(this,oi))),H(this,li)!==void 0&&H(this,U)&&this.media&&setTimeout(()=>{var e;this.media?.textTracks?.length&&((e=H(this,U))==null||e.dispatch({type:v.MEDIA_TOGGLE_SUBTITLES_REQUEST,detail:H(this,li)}))},0),this.hasAttribute(W.NO_HOTKEYS)?this.disableHotkeys():this.enableHotkeys()}disconnectedCallback(){var e,t,n,r,i;if((e=super.disconnectedCallback)==null||e.call(this),this.disableHotkeys(),H(this,U)){let e=H(this,U).getState();ti(this,li,!!e.mediaSubtitlesShowing?.length),(t=H(this,U))==null||t.dispatch({type:`fullscreenelementchangerequest`,detail:void 0}),(n=H(this,U))==null||n.dispatch({type:`documentelementchangerequest`,detail:void 0}),(r=H(this,U))==null||r.dispatch({type:v.MEDIA_TOGGLE_SUBTITLES_REQUEST,detail:!1})}H(this,si)&&((i=H(this,si))==null||i.call(this),ti(this,si,void 0)),this.unassociateElement(this),H(this,ai)&&(H(this,ai).remove(),ti(this,ai,null))}mediaSetCallback(e){var t;super.mediaSetCallback(e),(t=H(this,U))==null||t.dispatch({type:`mediaelementchangerequest`,detail:e}),e.hasAttribute(`tabindex`)||(e.tabIndex=-1)}mediaUnsetCallback(e){var t;super.mediaUnsetCallback(e),(t=H(this,U))==null||t.dispatch({type:`mediaelementchangerequest`,detail:void 0})}propagateMediaState(e,t){Ii(this.mediaStateReceivers,e,t)}associateElement(e){if(!e)return;let{associatedElementSubscriptions:t}=this;if(t.has(e))return;let n=Li(e,this.registerMediaStateReceiver.bind(this),this.unregisterMediaStateReceiver.bind(this));Object.values(v).forEach(t=>{e.addEventListener(t,H(this,ci))}),t.set(e,n)}unassociateElement(e){if(!e)return;let{associatedElementSubscriptions:t}=this;t.has(e)&&(t.get(e)(),t.delete(e),Object.values(v).forEach(t=>{e.removeEventListener(t,H(this,ci))}))}registerMediaStateReceiver(e){if(!e)return;let t=this.mediaStateReceivers;t.indexOf(e)>-1||(t.push(e),H(this,U)&&Object.entries(H(this,U).getState()).forEach(([t,n])=>{Ii([e],t,n)}))}unregisterMediaStateReceiver(e){let t=this.mediaStateReceivers,n=t.indexOf(e);n<0||t.splice(n,1)}enableHotkeys(){this.addEventListener(`keydown`,ni(this,pi,mi))}disableHotkeys(){this.removeEventListener(`keydown`,ni(this,pi,mi)),this.removeEventListener(`keyup`,H(this,fi))}get hotkeys(){return H(this,ri)}set hotkeys(e){I(this,W.HOTKEYS,e)}keyboardShortcutHandler(e){let t=e.target;if((t.getAttribute(W.KEYS_USED)?.split(` `)??t?.keysUsed??[]).map(e=>e===`Space`?` `:e).filter(Boolean).includes(e.key))return;let n,r,i;if(!H(this,ri).contains(Kn(e.key))&&(!e.shiftKey||e.key!==`/`&&e.key!==`?`||!H(this,ri).contains(`noshift+/`)))switch(e.key){case` `:case`k`:n=H(this,U).getState().mediaPaused?v.MEDIA_PLAY_REQUEST:v.MEDIA_PAUSE_REQUEST,this.dispatchEvent(new O.CustomEvent(n,{composed:!0,bubbles:!0}));break;case`m`:n=this.mediaStore.getState().mediaVolumeLevel===`off`?v.MEDIA_UNMUTE_REQUEST:v.MEDIA_MUTE_REQUEST,this.dispatchEvent(new O.CustomEvent(n,{composed:!0,bubbles:!0}));break;case`f`:n=this.mediaStore.getState().mediaIsFullscreen?v.MEDIA_EXIT_FULLSCREEN_REQUEST:v.MEDIA_ENTER_FULLSCREEN_REQUEST,this.dispatchEvent(new O.CustomEvent(n,{composed:!0,bubbles:!0}));break;case`c`:this.dispatchEvent(new O.CustomEvent(v.MEDIA_TOGGLE_SUBTITLES_REQUEST,{composed:!0,bubbles:!0}));break;case`ArrowLeft`:case`j`:{let e=this.hasAttribute(W.KEYBOARD_BACKWARD_SEEK_OFFSET)?+this.getAttribute(W.KEYBOARD_BACKWARD_SEEK_OFFSET):bi;r=Math.max((this.mediaStore.getState().mediaCurrentTime??0)-e,0),i=new O.CustomEvent(v.MEDIA_SEEK_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`ArrowRight`:case`l`:{let e=this.hasAttribute(W.KEYBOARD_FORWARD_SEEK_OFFSET)?+this.getAttribute(W.KEYBOARD_FORWARD_SEEK_OFFSET):bi;r=Math.max((this.mediaStore.getState().mediaCurrentTime??0)+e,0),i=new O.CustomEvent(v.MEDIA_SEEK_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`ArrowUp`:{let e=this.hasAttribute(W.KEYBOARD_UP_VOLUME_STEP)?+this.getAttribute(W.KEYBOARD_UP_VOLUME_STEP):xi;r=Math.min((this.mediaStore.getState().mediaVolume??1)+e,1),i=new O.CustomEvent(v.MEDIA_VOLUME_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`ArrowDown`:{let e=this.hasAttribute(W.KEYBOARD_DOWN_VOLUME_STEP)?+this.getAttribute(W.KEYBOARD_DOWN_VOLUME_STEP):xi;r=Math.max((this.mediaStore.getState().mediaVolume??1)-e,0),i=new O.CustomEvent(v.MEDIA_VOLUME_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`<`:{let e=this.mediaStore.getState().mediaPlaybackRate??1;r=Math.max(e-Si,Ci).toFixed(2),i=new O.CustomEvent(v.MEDIA_PLAYBACK_RATE_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`>`:{let e=this.mediaStore.getState().mediaPlaybackRate??1;r=Math.min(e+Si,wi).toFixed(2),i=new O.CustomEvent(v.MEDIA_PLAYBACK_RATE_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`/`:case`?`:e.shiftKey&&ni(this,_i,vi).call(this);break;case`p`:n=this.mediaStore.getState().mediaIsPip?v.MEDIA_EXIT_PIP_REQUEST:v.MEDIA_ENTER_PIP_REQUEST,i=new O.CustomEvent(n,{composed:!0,bubbles:!0}),this.dispatchEvent(i)}}};ri=new WeakMap,ii=new WeakMap,U=new WeakMap,ai=new WeakMap,oi=new WeakMap,si=new WeakMap,ci=new WeakMap,li=new WeakMap,ui=new WeakSet,di=function(){this.mediaStore=Qr({media:this.media,fullscreenElement:this.fullscreenElement,options:{defaultSubtitles:this.hasAttribute(W.DEFAULT_SUBTITLES),defaultDuration:this.hasAttribute(W.DEFAULT_DURATION)?+this.getAttribute(W.DEFAULT_DURATION):void 0,defaultStreamType:this.getAttribute(W.DEFAULT_STREAM_TYPE)??void 0,liveEdgeOffset:this.hasAttribute(W.LIVE_EDGE_OFFSET)?+this.getAttribute(W.LIVE_EDGE_OFFSET):void 0,seekToLiveOffset:this.hasAttribute(W.SEEK_TO_LIVE_OFFSET)?+this.getAttribute(W.SEEK_TO_LIVE_OFFSET):this.hasAttribute(W.LIVE_EDGE_OFFSET)?+this.getAttribute(W.LIVE_EDGE_OFFSET):void 0,noAutoSeekToLive:this.hasAttribute(W.NO_AUTO_SEEK_TO_LIVE),noVolumePref:this.hasAttribute(W.NO_VOLUME_PREF),noMutedPref:this.hasAttribute(W.NO_MUTED_PREF),noSubtitlesLangPref:this.hasAttribute(W.NO_SUBTITLES_LANG_PREF)}})},fi=new WeakMap,pi=new WeakSet,mi=function(e){let{metaKey:t,altKey:n,key:r,shiftKey:i}=e,a=i&&(r===`/`||r===`?`);if(a&&H(this,ai)?.open||t||n||!a&&!yi.includes(r)){this.removeEventListener(`keyup`,H(this,fi));return}let o=e.target,s=o instanceof HTMLElement&&(o.tagName.toLowerCase()===`media-volume-range`||o.tagName.toLowerCase()===`media-time-range`);[` `,`ArrowLeft`,`ArrowRight`,`ArrowUp`,`ArrowDown`].includes(r)&&!H(this,ri).contains(Kn(r))&&!s&&e.preventDefault(),this.addEventListener(`keyup`,H(this,fi),{once:!0})},hi=new WeakSet,gi=function(){H(this,ai)&&(H(this,ai).hotkeys=H(this,ri).value||void 0,(this.noHotkeys||H(this,ri).contains(`noshift+/`))&&(H(this,ai).open=!1))},_i=new WeakSet,vi=function(){H(this,ai)||(ti(this,ai,k.createElement(`media-keyboard-shortcuts-dialog`)),ni(this,hi,gi).call(this),this.appendChild(H(this,ai))),H(this,ai).open=!0};var Ei=Object.values(b),Di=Object.values(Ne),Oi=e=>{var t;let{observedAttributes:n}=e.constructor;!n&&e.nodeName?.includes(`-`)&&(O.customElements.upgrade(e),{observedAttributes:n}=e.constructor);let r=((t=(e?.getAttribute)?.call(e,y.MEDIA_CHROME_ATTRIBUTES))?.split)?.call(t,/\s+/);return Array.isArray(n||r)?(n||r).filter(e=>Ei.includes(e)):[]},ki=e=>(e.nodeName?.includes(`-`)&&O.customElements.get(e.nodeName?.toLowerCase())&&!(e instanceof O.customElements.get(e.nodeName.toLowerCase()))&&O.customElements.upgrade(e),Di.some(t=>t in e)),Ai=e=>ki(e)||!!Oi(e).length,ji=e=>(e?.join)?.call(e,`:`),Mi={[b.MEDIA_SUBTITLES_LIST]:pr,[b.MEDIA_SUBTITLES_SHOWING]:pr,[b.MEDIA_SEEKABLE]:ji,[b.MEDIA_BUFFERED]:e=>e?.map(ji).join(` `),[b.MEDIA_PREVIEW_COORDS]:e=>e?.join(` `),[b.MEDIA_RENDITION_LIST]:Re,[b.MEDIA_AUDIO_TRACK_LIST]:Be},Ni=async(e,t,n)=>{if(e.isConnected||await Ue(0),typeof n==`boolean`||n==null)return P(e,t,n);if(typeof n==`number`)return M(e,t,n);if(typeof n==`string`)return I(e,t,n);if(Array.isArray(n)&&!n.length)return e.removeAttribute(t);let r=Mi[t]?.call(Mi,n)??n;return e.setAttribute(t,r)},Pi=e=>!!e.closest?.call(e,`*[slot="media"]`),Fi=(e,t)=>{if(Pi(e))return;let n=(e,t)=>{Ai(e)&&t(e);let{children:n=[]}=e??{},r=e?.shadowRoot?.children??[];[...n,...r].forEach(e=>Fi(e,t))},r=e?.nodeName.toLowerCase();r.includes(`-`)&&!Ai(e)?O.customElements.whenDefined(r).then(()=>{n(e,t)}):n(e,t)},Ii=(e,t,n)=>{e.forEach(e=>{if(t in e){e[t]=n;return}let r=Oi(e),i=t.toLowerCase();r.includes(i)&&Ni(e,i,n)})},Li=(e,t,n)=>{Fi(e,t);let r=e=>{t(e?.composedPath()[0]??e.target)},i=e=>{n(e?.composedPath()[0]??e.target)};e.addEventListener(v.REGISTER_MEDIA_STATE_RECEIVER,r),e.addEventListener(v.UNREGISTER_MEDIA_STATE_RECEIVER,i);let a=e=>{e.forEach(e=>{let{addedNodes:r=[],removedNodes:i=[],type:a,target:o,attributeName:s}=e;a===`childList`?(Array.prototype.forEach.call(r,e=>Fi(e,t)),Array.prototype.forEach.call(i,e=>Fi(e,n))):a===`attributes`&&s===y.MEDIA_CHROME_ATTRIBUTES&&(Ai(o)?t(o):n(o))})},o=[],s=e=>{let r=e.target;r.name!==`media`&&(o.forEach(e=>Fi(e,n)),o=[...r.assignedElements({flatten:!0})],o.forEach(e=>Fi(e,t)))};e.addEventListener(`slotchange`,s);let c=new MutationObserver(a);return c.observe(e,{childList:!0,attributes:!0,subtree:!0}),()=>{Fi(e,n),e.removeEventListener(`slotchange`,s),c.disconnect(),e.removeEventListener(v.REGISTER_MEDIA_STATE_RECEIVER,r),e.removeEventListener(v.UNREGISTER_MEDIA_STATE_RECEIVER,i)}};O.customElements.get(`media-controller`)||O.customElements.define(`media-controller`,Ti);var Ri=Ti,zi={PLACEMENT:`placement`,BOUNDS:`bounds`};function Bi(e){return`
    <style>
      :host {
        --_tooltip-background-color: var(--media-tooltip-background-color, var(--media-secondary-color, rgba(20, 20, 30, .7)));
        --_tooltip-background: var(--media-tooltip-background, var(--_tooltip-background-color));
        --_tooltip-arrow-half-width: calc(var(--media-tooltip-arrow-width, 12px) / 2);
        --_tooltip-arrow-height: var(--media-tooltip-arrow-height, 5px);
        --_tooltip-arrow-background: var(--media-tooltip-arrow-color, var(--_tooltip-background-color));
        position: relative;
        pointer-events: none;
        display: var(--media-tooltip-display, inline-flex);
        justify-content: center;
        align-items: center;
        box-sizing: border-box;
        z-index: var(--media-tooltip-z-index, 1);
        background: var(--_tooltip-background);
        color: var(--media-text-color, var(--media-primary-color, rgb(238 238 238)));
        font: var(--media-font,
          var(--media-font-weight, 400)
          var(--media-font-size, 13px) /
          var(--media-text-content-height, var(--media-control-height, 18px))
          var(--media-font-family, helvetica neue, segoe ui, roboto, arial, sans-serif));
        padding: var(--media-tooltip-padding, .35em .7em);
        border: var(--media-tooltip-border, none);
        border-radius: var(--media-tooltip-border-radius, 5px);
        filter: var(--media-tooltip-filter, drop-shadow(0 0 4px rgba(0, 0, 0, .2)));
        white-space: var(--media-tooltip-white-space, nowrap);
      }

      :host([hidden]) {
        display: none;
      }

      img, svg {
        display: inline-block;
      }

      #arrow {
        position: absolute;
        width: 0px;
        height: 0px;
        border-style: solid;
        display: var(--media-tooltip-arrow-display, block);
      }

      :host(:not([placement])),
      :host([placement="top"]) {
        position: absolute;
        bottom: calc(100% + var(--media-tooltip-distance, 12px));
        left: 50%;
        transform: translate(calc(-50% - var(--media-tooltip-offset-x, 0px)), 0);
      }
      :host(:not([placement])) #arrow,
      :host([placement="top"]) #arrow {
        top: 100%;
        left: 50%;
        border-width: var(--_tooltip-arrow-height) var(--_tooltip-arrow-half-width) 0 var(--_tooltip-arrow-half-width);
        border-color: var(--_tooltip-arrow-background) transparent transparent transparent;
        transform: translate(calc(-50% + var(--media-tooltip-offset-x, 0px)), 0);
      }

      :host([placement="right"]) {
        position: absolute;
        left: calc(100% + var(--media-tooltip-distance, 12px));
        top: 50%;
        transform: translate(0, -50%);
      }
      :host([placement="right"]) #arrow {
        top: 50%;
        right: 100%;
        border-width: var(--_tooltip-arrow-half-width) var(--_tooltip-arrow-height) var(--_tooltip-arrow-half-width) 0;
        border-color: transparent var(--_tooltip-arrow-background) transparent transparent;
        transform: translate(0, -50%);
      }

      :host([placement="bottom"]) {
        position: absolute;
        top: calc(100% + var(--media-tooltip-distance, 12px));
        left: 50%;
        transform: translate(calc(-50% - var(--media-tooltip-offset-x, 0px)), 0);
      }
      :host([placement="bottom"]) #arrow {
        bottom: 100%;
        left: 50%;
        border-width: 0 var(--_tooltip-arrow-half-width) var(--_tooltip-arrow-height) var(--_tooltip-arrow-half-width);
        border-color: transparent transparent var(--_tooltip-arrow-background) transparent;
        transform: translate(calc(-50% + var(--media-tooltip-offset-x, 0px)), 0);
      }

      :host([placement="left"]) {
        position: absolute;
        right: calc(100% + var(--media-tooltip-distance, 12px));
        top: 50%;
        transform: translate(0, -50%);
      }
      :host([placement="left"]) #arrow {
        top: 50%;
        left: 100%;
        border-width: var(--_tooltip-arrow-half-width) 0 var(--_tooltip-arrow-half-width) var(--_tooltip-arrow-height);
        border-color: transparent transparent transparent var(--_tooltip-arrow-background);
        transform: translate(0, -50%);
      }
      
      :host([placement="none"]) #arrow {
        display: none;
      }
    </style>
    <slot></slot>
    <div id="arrow"></div>
  `}var Vi=class extends O.HTMLElement{constructor(){if(super(),this.updateXOffset=()=>{if(!Ct(this,{checkOpacity:!1,checkVisibilityCSS:!1}))return;let e=this.placement;if(e===`left`||e===`right`){this.style.removeProperty(`--media-tooltip-offset-x`);return}let t=getComputedStyle(this),n=bt(this,`#`+this.bounds)??mt(this);if(!n)return;let{x:r,width:i}=n.getBoundingClientRect(),{x:a,width:o}=this.getBoundingClientRect(),s=a+o,c=r+i,l=t.getPropertyValue(`--media-tooltip-offset-x`),u=l?parseFloat(l.replace(`px`,``)):0,d=t.getPropertyValue(`--media-tooltip-container-margin`),f=d?parseFloat(d.replace(`px`,``)):0,p=a-r+u-f,m=s-c+u+f;p<0?this.style.setProperty(`--media-tooltip-offset-x`,`${p}px`):m>0?this.style.setProperty(`--media-tooltip-offset-x`,`${m}px`):this.style.removeProperty(`--media-tooltip-offset-x`)},!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}if(this.arrowEl=this.shadowRoot.querySelector(`#arrow`),Object.prototype.hasOwnProperty.call(this,`placement`)){let e=this.placement;delete this.placement,this.placement=e}}static get observedAttributes(){return[zi.PLACEMENT,zi.BOUNDS]}get placement(){return F(this,zi.PLACEMENT)}set placement(e){I(this,zi.PLACEMENT,e)}get bounds(){return F(this,zi.BOUNDS)}set bounds(e){I(this,zi.BOUNDS,e)}};Vi.shadowRootOptions={mode:`open`},Vi.getTemplateHTML=Bi,O.customElements.get(`media-tooltip`)||O.customElements.define(`media-tooltip`,Vi);var Hi=Vi,Ui=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},G=(e,t,n)=>(Ui(e,t,`read from private field`),n?n.call(e):t.get(e)),Wi=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Gi=(e,t,n,r)=>(Ui(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),Ki=(e,t,n)=>(Ui(e,t,`access private method`),n),qi,Ji,Yi,Xi,Zi,Qi,$i,ea,ta,na={TOOLTIP_PLACEMENT:`tooltipplacement`,DISABLED:`disabled`,NO_TOOLTIP:`notooltip`};function ra(e,t={}){return`
    <style>
      :host {
        position: relative;
        font: var(--media-font,
          var(--media-font-weight, bold)
          var(--media-font-size, 14px) /
          var(--media-text-content-height, var(--media-control-height, 24px))
          var(--media-font-family, helvetica neue, segoe ui, roboto, arial, sans-serif));
        color: var(--media-text-color, var(--media-primary-color, rgb(238 238 238)));
        background: var(--media-control-background, var(--media-secondary-color, rgb(20 20 30 / .7)));
        padding: var(--media-button-padding, var(--media-control-padding, 10px));
        justify-content: var(--media-button-justify-content, center);
        display: inline-flex;
        align-items: center;
        vertical-align: middle;
        box-sizing: border-box;
        transition: background .15s linear;
        pointer-events: auto;
        cursor: var(--media-cursor, pointer);
        -webkit-tap-highlight-color: transparent;
      }

      
      :host(:focus-visible) {
        box-shadow: var(--media-focus-box-shadow, inset 0 0 0 2px rgb(27 127 204 / .9));
        outline: 0;
      }
      
      :host(:where(:focus)) {
        box-shadow: none;
        outline: 0;
      }

      :host(:hover) {
        background: var(--media-control-hover-background, rgba(50 50 70 / .7));
      }

      slot[name="icon"] {
        display: inline-flex;
        align-items: center;
      }

      svg, img, ::slotted(svg), ::slotted(img) {
        width: var(--media-button-icon-width);
        height: var(--media-button-icon-height, var(--media-control-height, 24px));
        transform: var(--media-button-icon-transform);
        transition: var(--media-button-icon-transition);
        fill: var(--media-icon-color, var(--media-primary-color, rgb(238 238 238)));
        vertical-align: middle;
        max-width: 100%;
        max-height: 100%;
        min-width: 100%;
      }

      media-tooltip {
        
        max-width: 0;
        overflow-x: clip;
        opacity: 0;
        transition: opacity .3s, max-width 0s 9s;
      }

      :host(:hover) media-tooltip,
      :host(:focus-visible) media-tooltip {
        max-width: 100vw;
        opacity: 1;
        transition: opacity .3s;
      }

      :host([notooltip]) slot[name="tooltip"] {
        display: none;
      }
    </style>

    ${this.getSlotTemplateHTML(e,t)}

    <slot name="tooltip">
      <media-tooltip part="tooltip" aria-hidden="true">
        <template shadowrootmode="${Hi.shadowRootOptions.mode}">
          ${Hi.getTemplateHTML({})}
        </template>
        <slot name="tooltip-content">
          ${this.getTooltipContentHTML(e)}
        </slot>
      </media-tooltip>
    </slot>
  `}function ia(e,t){return`
    <slot></slot>
  `}function aa(){return``}var K=class extends O.HTMLElement{constructor(){if(super(),Wi(this,Yi),Wi(this,ea),Wi(this,qi,void 0),this.preventClick=!1,this.tooltipEl=null,Wi(this,Ji,e=>{this.preventClick||this.handleClick(e),setTimeout(G(this,Zi),0)}),Wi(this,Zi,()=>{var e,t;(t=(e=this.tooltipEl)?.updateXOffset)==null||t.call(e)}),Wi(this,Qi,e=>{let{key:t}=e;this.keysUsed.includes(t)?this.preventClick||this.handleClick(e):this.removeEventListener(`keyup`,G(this,Qi))}),Wi(this,$i,e=>{let{metaKey:t,altKey:n,key:r}=e;t||n||!this.keysUsed.includes(r)?this.removeEventListener(`keyup`,G(this,Qi)):this.addEventListener(`keyup`,G(this,Qi),{once:!0})}),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes),t=this.constructor.getTemplateHTML(e);this.shadowRoot.setHTMLUnsafe?this.shadowRoot.setHTMLUnsafe(t):this.shadowRoot.innerHTML=t}this.tooltipEl=this.shadowRoot.querySelector(`media-tooltip`)}static get observedAttributes(){return[`disabled`,na.TOOLTIP_PLACEMENT,y.MEDIA_CONTROLLER,b.MEDIA_LANG]}enable(){this.addEventListener(`click`,G(this,Ji)),this.addEventListener(`keydown`,G(this,$i)),this.tabIndex=0}disable(){this.removeEventListener(`click`,G(this,Ji)),this.removeEventListener(`keydown`,G(this,$i)),this.removeEventListener(`keyup`,G(this,Qi)),this.tabIndex=-1}attributeChangedCallback(e,t,n){var r,i,a,o;e===y.MEDIA_CONTROLLER?(t&&((i=(r=G(this,qi))?.unassociateElement)==null||i.call(r,this),Gi(this,qi,null)),n&&this.isConnected&&(Gi(this,qi,this.getRootNode()?.getElementById(n)),(o=(a=G(this,qi))?.associateElement)==null||o.call(a,this))):e===`disabled`&&n!==t?n==null?this.enable():this.disable():e===na.TOOLTIP_PLACEMENT&&this.tooltipEl&&n!==t?this.tooltipEl.placement=n:e===b.MEDIA_LANG&&(this.shadowRoot.querySelector(`slot[name="tooltip-content"]`).innerHTML=this.constructor.getTooltipContentHTML()),Ki(this,Yi,Xi).call(this)&&G(this,Zi).call(this)}connectedCallback(){var e,t;let{style:n}=A(this.shadowRoot,`:host`);n.setProperty(`display`,`var(--media-control-display, var(--${this.localName}-display, inline-flex))`),this.hasAttribute(`disabled`)?this.disable():this.enable(),this.setAttribute(`role`,`button`);let r=this.getAttribute(y.MEDIA_CONTROLLER);r&&(Gi(this,qi,this.getRootNode()?.getElementById(r)),(t=(e=G(this,qi))?.associateElement)==null||t.call(e,this)),O.customElements.whenDefined(`media-tooltip`).then(()=>Ki(this,ea,ta).call(this))}disconnectedCallback(){var e,t;this.disable(),(t=(e=G(this,qi))?.unassociateElement)==null||t.call(e,this),Gi(this,qi,null),this.removeEventListener(`mouseenter`,G(this,Zi)),this.removeEventListener(`focus`,G(this,Zi)),this.removeEventListener(`click`,G(this,Ji))}get keysUsed(){return[`Enter`,` `]}get tooltipPlacement(){return F(this,na.TOOLTIP_PLACEMENT)}set tooltipPlacement(e){I(this,na.TOOLTIP_PLACEMENT,e)}get mediaController(){return F(this,y.MEDIA_CONTROLLER)}set mediaController(e){I(this,y.MEDIA_CONTROLLER,e)}get disabled(){return N(this,na.DISABLED)}set disabled(e){P(this,na.DISABLED,e)}get noTooltip(){return N(this,na.NO_TOOLTIP)}set noTooltip(e){P(this,na.NO_TOOLTIP,e)}handleClick(e){}};qi=new WeakMap,Ji=new WeakMap,Yi=new WeakSet,Xi=function(){try{return this.matches(`:hover, :focus-visible`)}catch{return this.matches(`:hover, :focus`)}},Zi=new WeakMap,Qi=new WeakMap,$i=new WeakMap,ea=new WeakSet,ta=function(){this.addEventListener(`mouseenter`,G(this,Zi)),this.addEventListener(`focus`,G(this,Zi)),this.addEventListener(`click`,G(this,Ji));let e=this.tooltipPlacement;e&&this.tooltipEl&&(this.tooltipEl.placement=e)},K.shadowRootOptions={mode:`open`},K.getTemplateHTML=ra,K.getSlotTemplateHTML=ia,K.getTooltipContentHTML=aa,O.customElements.get(`media-chrome-button`)||O.customElements.define(`media-chrome-button`,K);var oa=K,sa=`<svg aria-hidden="true" viewBox="0 0 26 24">
  <path d="M22.13 3H3.87a.87.87 0 0 0-.87.87v13.26a.87.87 0 0 0 .87.87h3.4L9 16H5V5h16v11h-4l1.72 2h3.4a.87.87 0 0 0 .87-.87V3.87a.87.87 0 0 0-.86-.87Zm-8.75 11.44a.5.5 0 0 0-.76 0l-4.91 5.73a.5.5 0 0 0 .38.83h9.82a.501.501 0 0 0 .38-.83l-4.91-5.73Z"/>
</svg>
`;function ca(e){return`
    <style>
      :host([${b.MEDIA_IS_AIRPLAYING}]) slot[name=icon] slot:not([name=exit]) {
        display: none !important;
      }

      
      :host(:not([${b.MEDIA_IS_AIRPLAYING}])) slot[name=icon] slot:not([name=enter]) {
        display: none !important;
      }

      :host([${b.MEDIA_IS_AIRPLAYING}]) slot[name=tooltip-enter],
      :host(:not([${b.MEDIA_IS_AIRPLAYING}])) slot[name=tooltip-exit] {
        display: none;
      }
    </style>

    <slot name="icon">
      <slot name="enter">${sa}</slot>
      <slot name="exit">${sa}</slot>
    </slot>
  `}function la(){return`
    <slot name="tooltip-enter">${E(`start airplay`)}</slot>
    <slot name="tooltip-exit">${E(`stop airplay`)}</slot>
  `}var ua=e=>{let t=e.mediaIsAirplaying?E(`stop airplay`):E(`start airplay`);e.setAttribute(`aria-label`,t)},da=class extends K{static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_IS_AIRPLAYING,b.MEDIA_AIRPLAY_UNAVAILABLE]}connectedCallback(){super.connectedCallback(),ua(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===b.MEDIA_IS_AIRPLAYING&&ua(this)}get mediaIsAirplaying(){return N(this,b.MEDIA_IS_AIRPLAYING)}set mediaIsAirplaying(e){P(this,b.MEDIA_IS_AIRPLAYING,e)}get mediaAirplayUnavailable(){return F(this,b.MEDIA_AIRPLAY_UNAVAILABLE)}set mediaAirplayUnavailable(e){I(this,b.MEDIA_AIRPLAY_UNAVAILABLE,e)}handleClick(){let e=new O.CustomEvent(v.MEDIA_AIRPLAY_REQUEST,{composed:!0,bubbles:!0});this.dispatchEvent(e)}};da.getSlotTemplateHTML=ca,da.getTooltipContentHTML=la,O.customElements.get(`media-airplay-button`)||O.customElements.define(`media-airplay-button`,da);var fa=da,pa=`<svg aria-hidden="true" viewBox="0 0 26 24">
  <path d="M22.83 5.68a2.58 2.58 0 0 0-2.3-2.5c-3.62-.24-11.44-.24-15.06 0a2.58 2.58 0 0 0-2.3 2.5c-.23 4.21-.23 8.43 0 12.64a2.58 2.58 0 0 0 2.3 2.5c3.62.24 11.44.24 15.06 0a2.58 2.58 0 0 0 2.3-2.5c.23-4.21.23-8.43 0-12.64Zm-11.39 9.45a3.07 3.07 0 0 1-1.91.57 3.06 3.06 0 0 1-2.34-1 3.75 3.75 0 0 1-.92-2.67 3.92 3.92 0 0 1 .92-2.77 3.18 3.18 0 0 1 2.43-1 2.94 2.94 0 0 1 2.13.78c.364.359.62.813.74 1.31l-1.43.35a1.49 1.49 0 0 0-1.51-1.17 1.61 1.61 0 0 0-1.29.58 2.79 2.79 0 0 0-.5 1.89 3 3 0 0 0 .49 1.93 1.61 1.61 0 0 0 1.27.58 1.48 1.48 0 0 0 1-.37 2.1 2.1 0 0 0 .59-1.14l1.4.44a3.23 3.23 0 0 1-1.07 1.69Zm7.22 0a3.07 3.07 0 0 1-1.91.57 3.06 3.06 0 0 1-2.34-1 3.75 3.75 0 0 1-.92-2.67 3.88 3.88 0 0 1 .93-2.77 3.14 3.14 0 0 1 2.42-1 3 3 0 0 1 2.16.82 2.8 2.8 0 0 1 .73 1.31l-1.43.35a1.49 1.49 0 0 0-1.51-1.21 1.61 1.61 0 0 0-1.29.58A2.79 2.79 0 0 0 15 12a3 3 0 0 0 .49 1.93 1.61 1.61 0 0 0 1.27.58 1.44 1.44 0 0 0 1-.37 2.1 2.1 0 0 0 .6-1.15l1.4.44a3.17 3.17 0 0 1-1.1 1.7Z"/>
</svg>`,ma=`<svg aria-hidden="true" viewBox="0 0 26 24">
  <path d="M17.73 14.09a1.4 1.4 0 0 1-1 .37 1.579 1.579 0 0 1-1.27-.58A3 3 0 0 1 15 12a2.8 2.8 0 0 1 .5-1.85 1.63 1.63 0 0 1 1.29-.57 1.47 1.47 0 0 1 1.51 1.2l1.43-.34A2.89 2.89 0 0 0 19 9.07a3 3 0 0 0-2.14-.78 3.14 3.14 0 0 0-2.42 1 3.91 3.91 0 0 0-.93 2.78 3.74 3.74 0 0 0 .92 2.66 3.07 3.07 0 0 0 2.34 1 3.07 3.07 0 0 0 1.91-.57 3.17 3.17 0 0 0 1.07-1.74l-1.4-.45c-.083.43-.3.822-.62 1.12Zm-7.22 0a1.43 1.43 0 0 1-1 .37 1.58 1.58 0 0 1-1.27-.58A3 3 0 0 1 7.76 12a2.8 2.8 0 0 1 .5-1.85 1.63 1.63 0 0 1 1.29-.57 1.47 1.47 0 0 1 1.51 1.2l1.43-.34a2.81 2.81 0 0 0-.74-1.32 2.94 2.94 0 0 0-2.13-.78 3.18 3.18 0 0 0-2.43 1 4 4 0 0 0-.92 2.78 3.74 3.74 0 0 0 .92 2.66 3.07 3.07 0 0 0 2.34 1 3.07 3.07 0 0 0 1.91-.57 3.23 3.23 0 0 0 1.07-1.74l-1.4-.45a2.06 2.06 0 0 1-.6 1.07Zm12.32-8.41a2.59 2.59 0 0 0-2.3-2.51C18.72 3.05 15.86 3 13 3c-2.86 0-5.72.05-7.53.17a2.59 2.59 0 0 0-2.3 2.51c-.23 4.207-.23 8.423 0 12.63a2.57 2.57 0 0 0 2.3 2.5c1.81.13 4.67.19 7.53.19 2.86 0 5.72-.06 7.53-.19a2.57 2.57 0 0 0 2.3-2.5c.23-4.207.23-8.423 0-12.63Zm-1.49 12.53a1.11 1.11 0 0 1-.91 1.11c-1.67.11-4.45.18-7.43.18-2.98 0-5.76-.07-7.43-.18a1.11 1.11 0 0 1-.91-1.11c-.21-4.14-.21-8.29 0-12.43a1.11 1.11 0 0 1 .91-1.11C7.24 4.56 10 4.49 13 4.49s5.76.07 7.43.18a1.11 1.11 0 0 1 .91 1.11c.21 4.14.21 8.29 0 12.43Z"/>
</svg>`;function ha(e){return`
    <style>
      :host([aria-checked="true"]) slot[name=off] {
        display: none !important;
      }

      
      :host(:not([aria-checked="true"])) slot[name=on] {
        display: none !important;
      }

      :host([aria-checked="true"]) slot[name=tooltip-enable],
      :host(:not([aria-checked="true"])) slot[name=tooltip-disable] {
        display: none;
      }
    </style>

    <slot name="icon">
      <slot name="on">${pa}</slot>
      <slot name="off">${ma}</slot>
    </slot>
  `}function ga(){return`
    <slot name="tooltip-enable">${E(`Enable captions`)}</slot>
    <slot name="tooltip-disable">${E(`Disable captions`)}</slot>
  `}var _a=e=>{e.setAttribute(`aria-checked`,vr(e).toString())},va=class extends K{static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_SUBTITLES_LIST,b.MEDIA_SUBTITLES_SHOWING]}connectedCallback(){super.connectedCallback(),this.setAttribute(`role`,`button`),this.setAttribute(`aria-label`,E(`closed captions`)),_a(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===b.MEDIA_SUBTITLES_SHOWING&&_a(this)}get mediaSubtitlesList(){return ya(this,b.MEDIA_SUBTITLES_LIST)}set mediaSubtitlesList(e){ba(this,b.MEDIA_SUBTITLES_LIST,e)}get mediaSubtitlesShowing(){return ya(this,b.MEDIA_SUBTITLES_SHOWING)}set mediaSubtitlesShowing(e){ba(this,b.MEDIA_SUBTITLES_SHOWING,e)}handleClick(){this.dispatchEvent(new O.CustomEvent(v.MEDIA_TOGGLE_SUBTITLES_REQUEST,{composed:!0,bubbles:!0}))}};va.getSlotTemplateHTML=ha,va.getTooltipContentHTML=ga;var ya=(e,t)=>{let n=e.getAttribute(t);return n?ur(n):[]},ba=(e,t,n)=>{if(!n?.length){e.removeAttribute(t);return}let r=pr(n);e.getAttribute(t)!==r&&e.setAttribute(t,r)};O.customElements.get(`media-captions-button`)||O.customElements.define(`media-captions-button`,va);var xa=va,Sa=`<svg aria-hidden="true" viewBox="0 0 24 24"><g><path class="cast_caf_icon_arch0" d="M1,18 L1,21 L4,21 C4,19.3 2.66,18 1,18 L1,18 Z"/><path class="cast_caf_icon_arch1" d="M1,14 L1,16 C3.76,16 6,18.2 6,21 L8,21 C8,17.13 4.87,14 1,14 L1,14 Z"/><path class="cast_caf_icon_arch2" d="M1,10 L1,12 C5.97,12 10,16.0 10,21 L12,21 C12,14.92 7.07,10 1,10 L1,10 Z"/><path class="cast_caf_icon_box" d="M21,3 L3,3 C1.9,3 1,3.9 1,5 L1,8 L3,8 L3,5 L21,5 L21,19 L14,19 L14,21 L21,21 C22.1,21 23,20.1 23,19 L23,5 C23,3.9 22.1,3 21,3 L21,3 Z"/></g></svg>`,Ca=`<svg aria-hidden="true" viewBox="0 0 24 24"><g><path class="cast_caf_icon_arch0" d="M1,18 L1,21 L4,21 C4,19.3 2.66,18 1,18 L1,18 Z"/><path class="cast_caf_icon_arch1" d="M1,14 L1,16 C3.76,16 6,18.2 6,21 L8,21 C8,17.13 4.87,14 1,14 L1,14 Z"/><path class="cast_caf_icon_arch2" d="M1,10 L1,12 C5.97,12 10,16.0 10,21 L12,21 C12,14.92 7.07,10 1,10 L1,10 Z"/><path class="cast_caf_icon_box" d="M21,3 L3,3 C1.9,3 1,3.9 1,5 L1,8 L3,8 L3,5 L21,5 L21,19 L14,19 L14,21 L21,21 C22.1,21 23,20.1 23,19 L23,5 C23,3.9 22.1,3 21,3 L21,3 Z"/><path class="cast_caf_icon_boxfill" d="M5,7 L5,8.63 C8,8.6 13.37,14 13.37,17 L19,17 L19,7 Z"/></g></svg>`;function wa(e){return`
    <style>
      :host([${b.MEDIA_IS_CASTING}]) slot[name=icon] slot:not([name=exit]) {
        display: none !important;
      }

      
      :host(:not([${b.MEDIA_IS_CASTING}])) slot[name=icon] slot:not([name=enter]) {
        display: none !important;
      }

      :host([${b.MEDIA_IS_CASTING}]) slot[name=tooltip-enter],
      :host(:not([${b.MEDIA_IS_CASTING}])) slot[name=tooltip-exit] {
        display: none;
      }
    </style>

    <slot name="icon">
      <slot name="enter">${Sa}</slot>
      <slot name="exit">${Ca}</slot>
    </slot>
  `}function Ta(){return`
    <slot name="tooltip-enter">${E(`Start casting`)}</slot>
    <slot name="tooltip-exit">${E(`Stop casting`)}</slot>
  `}var Ea=e=>{let t=e.mediaIsCasting?E(`stop casting`):E(`start casting`);e.setAttribute(`aria-label`,t)},Da=class extends K{static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_IS_CASTING,b.MEDIA_CAST_UNAVAILABLE]}connectedCallback(){super.connectedCallback(),Ea(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===b.MEDIA_IS_CASTING&&Ea(this)}get mediaIsCasting(){return N(this,b.MEDIA_IS_CASTING)}set mediaIsCasting(e){P(this,b.MEDIA_IS_CASTING,e)}get mediaCastUnavailable(){return F(this,b.MEDIA_CAST_UNAVAILABLE)}set mediaCastUnavailable(e){I(this,b.MEDIA_CAST_UNAVAILABLE,e)}handleClick(){let e=this.mediaIsCasting?v.MEDIA_EXIT_CAST_REQUEST:v.MEDIA_ENTER_CAST_REQUEST;this.dispatchEvent(new O.CustomEvent(e,{composed:!0,bubbles:!0}))}};Da.getSlotTemplateHTML=wa,Da.getTooltipContentHTML=Ta,O.customElements.get(`media-cast-button`)||O.customElements.define(`media-cast-button`,Da);var Oa=Da,ka=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},q=(e,t,n)=>(ka(e,t,`read from private field`),n?n.call(e):t.get(e)),J=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Aa=(e,t,n,r)=>(ka(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),ja=(e,t,n)=>(ka(e,t,`access private method`),n),Ma,Na,Pa,Fa,Ia,La,Ra,za,Ba,Va,Ha,Ua,Wa,Ga,Ka,qa,Ja,Ya,Xa,Za,Qa,$a,eo,to,no;function ro(e){return`
    <style>
      :host {
        --_focus-box-shadow: var(--media-focus-box-shadow, inset 0 0 0 2px rgb(27 127 204 / .9));
        --_media-range-padding: var(--media-range-padding, var(--media-control-padding, 10px));

        box-shadow: var(--_focus-visible-box-shadow, none);
        background: var(--media-control-background, var(--media-secondary-color, rgb(20 20 30 / .7)));
        height: calc(var(--media-control-height, 24px) + 2 * var(--_media-range-padding));
        display: inline-flex;
        align-items: center;
        
        vertical-align: middle;
        box-sizing: border-box;
        position: relative;
        width: 100px;
        transition: background .15s linear;
        cursor: var(--media-cursor, pointer);
        pointer-events: auto;
        touch-action: none; 
      }

      
      input[type=range]:focus {
        outline: 0;
      }
      input[type=range]:focus::-webkit-slider-runnable-track {
        outline: 0;
      }

      :host(:hover) {
        background: var(--media-control-hover-background, rgb(50 50 70 / .7));
      }

      #leftgap {
        padding-left: var(--media-range-padding-left, var(--_media-range-padding));
      }

      #rightgap {
        padding-right: var(--media-range-padding-right, var(--_media-range-padding));
      }

      #startpoint,
      #endpoint {
        position: absolute;
      }

      #endpoint {
        right: 0;
      }

      #container {
        
        width: var(--media-range-track-width, 100%);
        transform: translate(var(--media-range-track-translate-x, 0px), var(--media-range-track-translate-y, 0px));
        position: relative;
        height: 100%;
        display: flex;
        align-items: center;
        min-width: 40px;
      }

      #range {
        
        display: var(--media-time-range-hover-display, block);
        bottom: var(--media-time-range-hover-bottom, 0);
        height: var(--media-time-range-hover-height, max(100% , 25px));
        width: 100%;
        position: absolute;
        cursor: var(--media-cursor, pointer);

        -webkit-appearance: none; 
        -webkit-tap-highlight-color: transparent;
        background: transparent; 
        margin: 0;
        z-index: 1;
      }

      @media (hover: hover) {
        #range {
          bottom: var(--media-time-range-hover-bottom, 0);
          height: var(--media-time-range-hover-height, max(100%, 20px));
        }
      }

      
      
      #range::-webkit-slider-thumb {
        -webkit-appearance: none;
        background: transparent;
        width: .1px;
        height: .1px;
      }

      
      #range::-moz-range-thumb {
        background: transparent;
        border: transparent;
        width: .1px;
        height: .1px;
      }

      #appearance {
        height: var(--media-range-track-height, 4px);
        display: flex;
        flex-direction: column;
        justify-content: center;
        width: 100%;
        position: absolute;
        
        will-change: transform;
      }

      #track {
        background: var(--media-range-track-background, rgb(255 255 255 / .2));
        border-radius: var(--media-range-track-border-radius, 1px);
        border: var(--media-range-track-border, none);
        outline: var(--media-range-track-outline);
        outline-offset: var(--media-range-track-outline-offset);
        backdrop-filter: var(--media-range-track-backdrop-filter);
        -webkit-backdrop-filter: var(--media-range-track-backdrop-filter);
        box-shadow: var(--media-range-track-box-shadow, none);
        position: absolute;
        width: 100%;
        height: 100%;
        overflow: hidden;
      }

      #progress,
      #pointer {
        position: absolute;
        height: 100%;
        will-change: width;
      }

      #progress {
        background: var(--media-range-bar-color, var(--media-primary-color, rgb(238 238 238)));
        transition: var(--media-range-track-transition);
      }

      #pointer {
        background: var(--media-range-track-pointer-background);
        border-right: var(--media-range-track-pointer-border-right);
        transition: visibility .25s, opacity .25s;
        visibility: hidden;
        opacity: 0;
      }

      @media (hover: hover) {
        :host(:hover) #pointer {
          transition: visibility .5s, opacity .5s;
          visibility: visible;
          opacity: 1;
        }
      }

      #thumb,
      ::slotted([slot=thumb]) {
        width: var(--media-range-thumb-width, 10px);
        height: var(--media-range-thumb-height, 10px);
        transition: var(--media-range-thumb-transition);
        transform: var(--media-range-thumb-transform, none);
        opacity: var(--media-range-thumb-opacity, 1);
        translate: -50%;
        position: absolute;
        left: 0;
        cursor: var(--media-cursor, pointer);
      }

      #thumb {
        border-radius: var(--media-range-thumb-border-radius, 10px);
        background: var(--media-range-thumb-background, var(--media-primary-color, rgb(238 238 238)));
        box-shadow: var(--media-range-thumb-box-shadow, 1px 1px 1px transparent);
        border: var(--media-range-thumb-border, none);
      }

      :host([disabled]) #thumb {
        background-color: #777;
      }

      .segments #appearance {
        height: var(--media-range-segment-hover-height, 7px);
      }

      #track {
        clip-path: url(#segments-clipping);
      }

      #segments {
        --segments-gap: var(--media-range-segments-gap, 2px);
        position: absolute;
        width: 100%;
        height: 100%;
      }

      #segments-clipping {
        transform: translateX(calc(var(--segments-gap) / 2));
      }

      #segments-clipping:empty {
        display: none;
      }

      #segments-clipping rect {
        height: var(--media-range-track-height, 4px);
        y: calc((var(--media-range-segment-hover-height, 7px) - var(--media-range-track-height, 4px)) / 2);
        transition: var(--media-range-segment-transition, transform .1s ease-in-out);
        transform: var(--media-range-segment-transform, scaleY(1));
        transform-origin: center;
      }

      /* Visible label for accessibility - positioned off-screen but technically visible (Firefox requires visible labels) */
      #range-label {
        position: absolute;
        left: -10000px;
        background: var(--media-control-background, var(--media-secondary-color, rgb(20 20 30 / .7)));
        pointer-events: none;
      }
    </style>
    <div id="leftgap"></div>
    <div id="container">
      <div id="startpoint"></div>
      <div id="endpoint"></div>
      <div id="appearance">
        <div id="track" part="track">
          <div id="pointer"></div>
          <div id="progress" part="progress"></div>
        </div>
        <slot name="thumb">
          <div id="thumb" part="thumb"></div>
        </slot>
        <svg id="segments" aria-hidden="true"><clipPath id="segments-clipping"></clipPath></svg>
      </div>
        <input id="range" type="range" min="0" max="1" step="any" value="0">
        <label for="range" id="range-label"></label>

      ${this.getContainerTemplateHTML(e)}
    </div>
    <div id="rightgap"></div>
  `}function io(e){return``}var ao=class extends O.HTMLElement{constructor(){if(super(),J(this,Va),J(this,Ua),J(this,Ga),J(this,qa),J(this,Ya),J(this,Za),J(this,$a),J(this,to),J(this,Ma,void 0),J(this,Na,void 0),J(this,Pa,void 0),J(this,Fa,void 0),J(this,Ia,{}),J(this,La,[]),J(this,Ra,()=>{if(this.range.matches(`:focus-visible`)){let{style:e}=A(this.shadowRoot,`:host`);e.setProperty(`--_focus-visible-box-shadow`,`var(--_focus-box-shadow)`)}}),J(this,za,()=>{let{style:e}=A(this.shadowRoot,`:host`);e.removeProperty(`--_focus-visible-box-shadow`)}),J(this,Ba,()=>{let e=this.shadowRoot.querySelector(`#segments-clipping`);e&&e.parentNode.append(e)}),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes),t=this.constructor.getTemplateHTML(e);this.shadowRoot.setHTMLUnsafe?this.shadowRoot.setHTMLUnsafe(t):this.shadowRoot.innerHTML=t}this.container=this.shadowRoot.querySelector(`#container`),Aa(this,Pa,this.shadowRoot.querySelector(`#startpoint`)),Aa(this,Fa,this.shadowRoot.querySelector(`#endpoint`)),this.range=this.shadowRoot.querySelector(`#range`),this.appearance=this.shadowRoot.querySelector(`#appearance`)}static get observedAttributes(){return[`disabled`,`aria-disabled`,y.MEDIA_CONTROLLER]}attributeChangedCallback(e,t,n){var r,i,a,o;e===y.MEDIA_CONTROLLER?(t&&((i=(r=q(this,Ma))?.unassociateElement)==null||i.call(r,this),Aa(this,Ma,null)),n&&this.isConnected&&(Aa(this,Ma,this.getRootNode()?.getElementById(n)),(o=(a=q(this,Ma))?.associateElement)==null||o.call(a,this))):(e===`disabled`||e===`aria-disabled`&&t!==n)&&(n==null?(this.range.removeAttribute(e),ja(this,Ua,Wa).call(this)):(this.range.setAttribute(e,n),ja(this,Ga,Ka).call(this)))}connectedCallback(){var e,t;let{style:n}=A(this.shadowRoot,`:host`);n.setProperty(`display`,`var(--media-control-display, var(--${this.localName}-display, inline-flex))`),q(this,Ia).pointer=A(this.shadowRoot,`#pointer`),q(this,Ia).progress=A(this.shadowRoot,`#progress`),q(this,Ia).thumb=A(this.shadowRoot,`#thumb, ::slotted([slot="thumb"])`),q(this,Ia).activeSegment=A(this.shadowRoot,`#segments-clipping rect:nth-child(0)`);let r=this.getAttribute(y.MEDIA_CONTROLLER);r&&(Aa(this,Ma,this.getRootNode()?.getElementById(r)),(t=(e=q(this,Ma))?.associateElement)==null||t.call(e,this)),this.updateBar(),this.shadowRoot.addEventListener(`focusin`,q(this,Ra)),this.shadowRoot.addEventListener(`focusout`,q(this,za)),ja(this,Ua,Wa).call(this),dt(this.container,q(this,Ba))}disconnectedCallback(){var e,t;ja(this,Ga,Ka).call(this),(t=(e=q(this,Ma))?.unassociateElement)==null||t.call(e,this),Aa(this,Ma,null),this.shadowRoot.removeEventListener(`focusin`,q(this,Ra)),this.shadowRoot.removeEventListener(`focusout`,q(this,za)),ft(this.container,q(this,Ba))}updatePointerBar(e){var t;(t=q(this,Ia).pointer)==null||t.style.setProperty(`width`,`${this.getPointerRatio(e)*100}%`)}updateBar(){var e,t;let n=this.range.valueAsNumber*100;(e=q(this,Ia).progress)==null||e.style.setProperty(`width`,`${n}%`),(t=q(this,Ia).thumb)==null||t.style.setProperty(`left`,`${n}%`)}updateSegments(e){let t=this.shadowRoot.querySelector(`#segments-clipping`);if(t.textContent=``,this.container.classList.toggle(`segments`,!!e?.length),!e?.length)return;let n=[...new Set([+this.range.min,...e.flatMap(e=>[e.start,e.end]),+this.range.max])];Aa(this,La,[...n]);let r=n.pop();for(let[e,i]of n.entries()){let[a,o]=[e===0,e===n.length-1],s=a?`calc(var(--segments-gap) / -1)`:`${i*100}%`,c=`calc(${((o?r:n[e+1])-i)*100}%${a||o?``:` - var(--segments-gap)`})`,l=k.createElementNS(`http://www.w3.org/2000/svg`,`rect`),u=Et(this.shadowRoot,`#segments-clipping rect:nth-child(${e+1})`);u.style.setProperty(`x`,s),u.style.setProperty(`width`,c),t.append(l)}}getPointerRatio(e){return wt(e.clientX,e.clientY,q(this,Pa).getBoundingClientRect(),q(this,Fa).getBoundingClientRect())}get dragging(){return this.hasAttribute(`dragging`)}handleEvent(e){switch(e.type){case`pointermove`:ja(this,to,no).call(this,e);break;case`input`:this.updateBar();break;case`pointerenter`:ja(this,Ya,Xa).call(this,e);break;case`pointerdown`:ja(this,qa,Ja).call(this,e);break;case`pointerup`:ja(this,Za,Qa).call(this);break;case`pointerleave`:ja(this,$a,eo).call(this)}}get keysUsed(){return[`ArrowUp`,`ArrowRight`,`ArrowDown`,`ArrowLeft`]}};Ma=new WeakMap,Na=new WeakMap,Pa=new WeakMap,Fa=new WeakMap,Ia=new WeakMap,La=new WeakMap,Ra=new WeakMap,za=new WeakMap,Ba=new WeakMap,Va=new WeakSet,Ha=function(e){let t=q(this,Ia).activeSegment;if(!t)return;let n=this.getPointerRatio(e),r=`#segments-clipping rect:nth-child(${q(this,La).findIndex((e,t,r)=>{let i=r[t+1];return i!=null&&n>=e&&n<=i})+1})`;t.selectorText==r&&t.style.transform||(t.selectorText=r,t.style.setProperty(`transform`,`var(--media-range-segment-hover-transform, scaleY(2))`))},Ua=new WeakSet,Wa=function(){!this.hasAttribute(`disabled`)&&this.isConnected&&(this.addEventListener(`input`,this),this.addEventListener(`pointerdown`,this),this.addEventListener(`pointerenter`,this))},Ga=new WeakSet,Ka=function(){var e,t;this.removeEventListener(`input`,this),this.removeEventListener(`pointerdown`,this),this.removeEventListener(`pointerenter`,this),this.removeEventListener(`pointerleave`,this),(e=O.window)==null||e.removeEventListener(`pointerup`,this),(t=O.window)==null||t.removeEventListener(`pointermove`,this)},qa=new WeakSet,Ja=function(e){var t;Aa(this,Na,e.composedPath().includes(this.range)),(t=O.window)==null||t.addEventListener(`pointerup`,this,{once:!0})},Ya=new WeakSet,Xa=function(e){var t;e.pointerType!==`mouse`&&ja(this,qa,Ja).call(this,e),this.addEventListener(`pointerleave`,this,{once:!0}),(t=O.window)==null||t.addEventListener(`pointermove`,this)},Za=new WeakSet,Qa=function(){var e;(e=O.window)==null||e.removeEventListener(`pointerup`,this),this.toggleAttribute(`dragging`,!1),this.range.disabled=this.hasAttribute(`disabled`)},$a=new WeakSet,eo=function(){var e,t;this.removeEventListener(`pointerleave`,this),(e=O.window)==null||e.removeEventListener(`pointermove`,this),this.toggleAttribute(`dragging`,!1),this.range.disabled=this.hasAttribute(`disabled`),(t=q(this,Ia).activeSegment)==null||t.style.removeProperty(`transform`)},to=new WeakSet,no=function(e){(e.pointerType!==`pen`||e.buttons!==0)&&(this.toggleAttribute(`dragging`,e.buttons===1||e.pointerType!==`mouse`),this.updatePointerBar(e),ja(this,Va,Ha).call(this,e),this.dragging&&(e.pointerType!==`mouse`||!q(this,Na))&&(this.range.disabled=!0,this.range.valueAsNumber=this.getPointerRatio(e),this.range.dispatchEvent(new Event(`input`,{bubbles:!0,composed:!0}))))},ao.shadowRootOptions={mode:`open`},ao.getTemplateHTML=ro,ao.getContainerTemplateHTML=io,O.customElements.get(`media-chrome-range`)||O.customElements.define(`media-chrome-range`,ao);var oo=ao,so=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},co=(e,t,n)=>(so(e,t,`read from private field`),n?n.call(e):t.get(e)),lo=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},uo=(e,t,n,r)=>(so(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),fo;function po(e){return`
    <style>
      :host {
        
        box-sizing: border-box;
        display: var(--media-control-display, var(--media-control-bar-display, inline-flex));
        color: var(--media-text-color, var(--media-primary-color, rgb(238 238 238)));
        --media-loading-indicator-icon-height: 44px;
      }

      ::slotted(media-time-range),
      ::slotted(media-volume-range) {
        min-height: 100%;
      }

      ::slotted(media-time-range),
      ::slotted(media-clip-selector) {
        flex-grow: 1;
      }

      ::slotted([role="menu"]) {
        position: absolute;
      }
    </style>

    <slot></slot>
  `}var mo=class extends O.HTMLElement{constructor(){if(super(),lo(this,fo,void 0),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}}static get observedAttributes(){return[y.MEDIA_CONTROLLER]}attributeChangedCallback(e,t,n){var r,i,a,o;e===y.MEDIA_CONTROLLER&&(t&&((i=(r=co(this,fo))?.unassociateElement)==null||i.call(r,this),uo(this,fo,null)),n&&this.isConnected&&(uo(this,fo,this.getRootNode()?.getElementById(n)),(o=(a=co(this,fo))?.associateElement)==null||o.call(a,this)))}connectedCallback(){var e,t;let n=this.getAttribute(y.MEDIA_CONTROLLER);n&&(uo(this,fo,this.getRootNode()?.getElementById(n)),(t=(e=co(this,fo))?.associateElement)==null||t.call(e,this))}disconnectedCallback(){var e,t;(t=(e=co(this,fo))?.unassociateElement)==null||t.call(e,this),uo(this,fo,null)}};fo=new WeakMap,mo.shadowRootOptions={mode:`open`},mo.getTemplateHTML=po,O.customElements.get(`media-control-bar`)||O.customElements.define(`media-control-bar`,mo);var ho=mo,go=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},_o=(e,t,n)=>(go(e,t,`read from private field`),n?n.call(e):t.get(e)),vo=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},yo=(e,t,n,r)=>(go(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),bo;function xo(e,t={}){return`
    <style>
      :host {
        font: var(--media-font,
          var(--media-font-weight, normal)
          var(--media-font-size, 14px) /
          var(--media-text-content-height, var(--media-control-height, 24px))
          var(--media-font-family, helvetica neue, segoe ui, roboto, arial, sans-serif));
        color: var(--media-text-color, var(--media-primary-color, rgb(238 238 238)));
        background: var(--media-text-background, var(--media-control-background, var(--media-secondary-color, rgb(20 20 30 / .7))));
        padding: var(--media-control-padding, 10px);
        display: inline-flex;
        justify-content: center;
        align-items: center;
        vertical-align: middle;
        box-sizing: border-box;
        text-align: center;
        pointer-events: auto;
      }

      
      :host(:focus-visible) {
        box-shadow: var(--media-focus-box-shadow, inset 0 0 0 2px rgb(27 127 204 / .9));
        outline: 0;
      }

      
      :host(:where(:focus)) {
        box-shadow: none;
        outline: 0;
      }
    </style>

    ${this.getSlotTemplateHTML(e,t)}
  `}function So(e,t){return`
    <slot></slot>
  `}var Co=class extends O.HTMLElement{constructor(){if(super(),vo(this,bo,void 0),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}}static get observedAttributes(){return[y.MEDIA_CONTROLLER]}attributeChangedCallback(e,t,n){var r,i,a,o;e===y.MEDIA_CONTROLLER&&(t&&((i=(r=_o(this,bo))?.unassociateElement)==null||i.call(r,this),yo(this,bo,null)),n&&this.isConnected&&(yo(this,bo,this.getRootNode()?.getElementById(n)),(o=(a=_o(this,bo))?.associateElement)==null||o.call(a,this)))}connectedCallback(){var e,t;let{style:n}=A(this.shadowRoot,`:host`);n.setProperty(`display`,`var(--media-control-display, var(--${this.localName}-display, inline-flex))`);let r=this.getAttribute(y.MEDIA_CONTROLLER);r&&(yo(this,bo,this.getRootNode()?.getElementById(r)),(t=(e=_o(this,bo))?.associateElement)==null||t.call(e,this))}disconnectedCallback(){var e,t;(t=(e=_o(this,bo))?.unassociateElement)==null||t.call(e,this),yo(this,bo,null)}};bo=new WeakMap,Co.shadowRootOptions={mode:`open`},Co.getTemplateHTML=xo,Co.getSlotTemplateHTML=So,O.customElements.get(`media-text-display`)||O.customElements.define(`media-text-display`,Co);var wo=Co,To=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Eo=(e,t,n)=>(To(e,t,`read from private field`),n?n.call(e):t.get(e)),Do=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Oo=(e,t,n,r)=>(To(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),ko;function Ao(e,t){return`
    <slot>${D(t.mediaDuration)}</slot>
  `}var jo=class extends Co{constructor(){super(),Do(this,ko,void 0),Oo(this,ko,this.shadowRoot.querySelector(`slot`)),Eo(this,ko).textContent=D(this.mediaDuration??0)}static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_DURATION]}attributeChangedCallback(e,t,n){e===b.MEDIA_DURATION&&(Eo(this,ko).textContent=D(+n)),super.attributeChangedCallback(e,t,n)}get mediaDuration(){return j(this,b.MEDIA_DURATION)}set mediaDuration(e){M(this,b.MEDIA_DURATION,e)}};ko=new WeakMap,jo.getSlotTemplateHTML=Ao,O.customElements.get(`media-duration-display`)||O.customElements.define(`media-duration-display`,jo);var Mo=jo,No={2:E(`Network Error`),3:E(`Decode Error`),4:E(`Source Not Supported`),5:E(`Encryption Error`)},Po={2:E(`A network error caused the media download to fail.`),3:E(`A media error caused playback to be aborted. The media could be corrupt or your browser does not support this format.`),4:E(`An unsupported error occurred. The server or network failed, or your browser does not support this format.`),5:E(`The media is encrypted and there are no keys to decrypt it.`)},Fo=e=>e.code===1?null:{title:No[e.code]??`Error ${e.code}`,message:Po[e.code]??e.message},Io=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Lo=(e,t,n)=>(Io(e,t,`read from private field`),n?n.call(e):t.get(e)),Ro=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},zo=(e,t,n,r)=>(Io(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),Bo;function Vo(e){let t=+e.mediaerrorcode;return`
    <style>
      :host {
        background: rgb(20 20 30 / .8);
      }

      #content {
        display: block;
        padding: 1.2em 1.5em;
      }

      h3,
      p {
        margin-block: 0 .3em;
      }
    </style>
    <slot name="error-${t}" id="content">
      ${Uo({code:t,message:e.mediaerrormessage})}
    </slot>
  `}function Ho(e){return e.code&&Fo(e)!==null}function Uo(e){let{title:t,message:n}=Fo(e)??{},r=Ge(e.code),i=``;return t&&(i+=`<slot name="error-${r}-title"><h3>${Ge(t)}</h3></slot>`),n&&(i+=`<slot name="error-${r}-message"><p>${Ge(n)}</p></slot>`),i}var Wo=[b.MEDIA_ERROR_CODE,b.MEDIA_ERROR_MESSAGE],Go=class extends Ln{constructor(){super(...arguments),Ro(this,Bo,null)}static get observedAttributes(){return[...super.observedAttributes,...Wo]}formatErrorMessage(e){return this.constructor.formatErrorMessage(e)}attributeChangedCallback(e,t,n){if(super.attributeChangedCallback(e,t,n),!Wo.includes(e))return;let r=this.mediaError??{code:this.mediaErrorCode,message:this.mediaErrorMessage};if(this.open=Ho(r),this.open&&(this.shadowRoot.querySelector(`slot`).name=`error-${this.mediaErrorCode}`,this.shadowRoot.querySelector(`#content`).innerHTML=this.formatErrorMessage(r),!this.hasAttribute(`aria-label`))){let{title:e}=Fo(r);e&&this.setAttribute(`aria-label`,e)}}get mediaError(){return Lo(this,Bo)}set mediaError(e){zo(this,Bo,e)}get mediaErrorCode(){return j(this,`mediaerrorcode`)}set mediaErrorCode(e){M(this,`mediaerrorcode`,e)}get mediaErrorMessage(){return F(this,`mediaerrormessage`)}set mediaErrorMessage(e){I(this,`mediaerrormessage`,e)}};Bo=new WeakMap,Go.getSlotTemplateHTML=Vo,Go.formatErrorMessage=Uo,O.customElements.get(`media-error-dialog`)||O.customElements.define(`media-error-dialog`,Go);var Ko=Go,qo=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Jo=(e,t,n)=>(qo(e,t,`read from private field`),n?n.call(e):t.get(e)),Yo=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Xo=(e,t,n,r)=>(qo(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),Zo,Qo=`<svg aria-hidden="true" viewBox="0 0 26 24">
  <path d="M16 3v2.5h3.5V9H22V3h-6ZM4 9h2.5V5.5H10V3H4v6Zm15.5 9.5H16V21h6v-6h-2.5v3.5ZM6.5 15H4v6h6v-2.5H6.5V15Z"/>
</svg>`,$o=`<svg aria-hidden="true" viewBox="0 0 26 24">
  <path d="M18.5 6.5V3H16v6h6V6.5h-3.5ZM16 21h2.5v-3.5H22V15h-6v6ZM4 17.5h3.5V21H10v-6H4v2.5Zm3.5-11H4V9h6V3H7.5v3.5Z"/>
</svg>`;function es(e){return`
    <style>
      :host([${b.MEDIA_IS_FULLSCREEN}]) slot[name=icon] slot:not([name=exit]) {
        display: none !important;
      }

      
      :host(:not([${b.MEDIA_IS_FULLSCREEN}])) slot[name=icon] slot:not([name=enter]) {
        display: none !important;
      }

      :host([${b.MEDIA_IS_FULLSCREEN}]) slot[name=tooltip-enter],
      :host(:not([${b.MEDIA_IS_FULLSCREEN}])) slot[name=tooltip-exit] {
        display: none;
      }
    </style>

    <slot name="icon">
      <slot name="enter">${Qo}</slot>
      <slot name="exit">${$o}</slot>
    </slot>
  `}function ts(){return`
    <slot name="tooltip-enter">${E(`Enter fullscreen mode`)}</slot>
    <slot name="tooltip-exit">${E(`Exit fullscreen mode`)}</slot>
  `}var ns=e=>{let t=e.mediaIsFullscreen?E(`exit fullscreen mode`):E(`enter fullscreen mode`);e.setAttribute(`aria-label`,t)},rs=class extends K{constructor(){super(...arguments),Yo(this,Zo,null)}static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_IS_FULLSCREEN,b.MEDIA_FULLSCREEN_UNAVAILABLE]}connectedCallback(){super.connectedCallback(),ns(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===b.MEDIA_IS_FULLSCREEN&&ns(this)}get mediaFullscreenUnavailable(){return F(this,b.MEDIA_FULLSCREEN_UNAVAILABLE)}set mediaFullscreenUnavailable(e){I(this,b.MEDIA_FULLSCREEN_UNAVAILABLE,e)}get mediaIsFullscreen(){return N(this,b.MEDIA_IS_FULLSCREEN)}set mediaIsFullscreen(e){P(this,b.MEDIA_IS_FULLSCREEN,e)}handleClick(e){Xo(this,Zo,e);let t=Jo(this,Zo)instanceof PointerEvent,n=this.mediaIsFullscreen?new O.CustomEvent(v.MEDIA_EXIT_FULLSCREEN_REQUEST,{composed:!0,bubbles:!0}):new O.CustomEvent(v.MEDIA_ENTER_FULLSCREEN_REQUEST,{composed:!0,bubbles:!0,detail:t});this.dispatchEvent(n)}};Zo=new WeakMap,rs.getSlotTemplateHTML=es,rs.getTooltipContentHTML=ts,O.customElements.get(`media-fullscreen-button`)||O.customElements.define(`media-fullscreen-button`,rs);var is=rs,{MEDIA_TIME_IS_LIVE:as,MEDIA_PAUSED:os}=b,{MEDIA_SEEK_TO_LIVE_REQUEST:ss,MEDIA_PLAY_REQUEST:cs}=v,ls=`<svg viewBox="0 0 6 12" aria-hidden="true"><circle cx="3" cy="6" r="2"></circle></svg>`;function us(e){return`
    <style>
      :host { --media-tooltip-display: none; }
      
      slot[name=indicator] > *,
      :host ::slotted([slot=indicator]) {
        
        min-width: auto;
        fill: var(--media-live-button-icon-color, rgb(140, 140, 140));
        color: var(--media-live-button-icon-color, rgb(140, 140, 140));
      }

      :host([${as}]:not([${os}])) slot[name=indicator] > *,
      :host([${as}]:not([${os}])) ::slotted([slot=indicator]) {
        fill: var(--media-live-button-indicator-color, rgb(255, 0, 0));
        color: var(--media-live-button-indicator-color, rgb(255, 0, 0));
      }

      :host([${as}]:not([${os}])) {
        cursor: var(--media-cursor, not-allowed);
      }

      slot[name=text]{
        text-transform: uppercase;
      }

    </style>

    <slot name="indicator">${ls}</slot>
    
    <slot name="spacer">&nbsp;</slot><slot name="text">${E(`live`)}</slot>
  `}var ds=e=>{let t=e.mediaPaused||!e.mediaTimeIsLive,n=E(t?`seek to live`:`playing live`);e.setAttribute(`aria-label`,n);let r=e.shadowRoot?.querySelector(`slot[name="text"]`);r&&(r.textContent=E(`live`)),t?e.removeAttribute(`aria-disabled`):e.setAttribute(`aria-disabled`,`true`)},fs=class extends K{static get observedAttributes(){return[...super.observedAttributes,as,os]}connectedCallback(){super.connectedCallback(),ds(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),ds(this)}get mediaPaused(){return N(this,b.MEDIA_PAUSED)}set mediaPaused(e){P(this,b.MEDIA_PAUSED,e)}get mediaTimeIsLive(){return N(this,b.MEDIA_TIME_IS_LIVE)}set mediaTimeIsLive(e){P(this,b.MEDIA_TIME_IS_LIVE,e)}handleClick(){(this.mediaPaused||!this.mediaTimeIsLive)&&(this.dispatchEvent(new O.CustomEvent(ss,{composed:!0,bubbles:!0})),this.hasAttribute(os)&&this.dispatchEvent(new O.CustomEvent(cs,{composed:!0,bubbles:!0})))}};fs.getSlotTemplateHTML=us,O.customElements.get(`media-live-button`)||O.customElements.define(`media-live-button`,fs);var ps=fs,ms=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},hs=(e,t,n)=>(ms(e,t,`read from private field`),n?n.call(e):t.get(e)),gs=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},_s=(e,t,n,r)=>(ms(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),vs,ys,bs={LOADING_DELAY:`loadingdelay`,NO_AUTOHIDE:`noautohide`},xs=500,Ss=`
<svg aria-hidden="true" viewBox="0 0 100 100">
  <path d="M73,50c0-12.7-10.3-23-23-23S27,37.3,27,50 M30.9,50c0-10.5,8.5-19.1,19.1-19.1S69.1,39.5,69.1,50">
    <animateTransform
       attributeName="transform"
       attributeType="XML"
       type="rotate"
       dur="1s"
       from="0 50 50"
       to="360 50 50"
       repeatCount="indefinite" />
  </path>
</svg>
`;function Cs(e){return`
    <style>
      :host {
        display: var(--media-control-display, var(--media-loading-indicator-display, inline-block));
        vertical-align: middle;
        box-sizing: border-box;
        --_loading-indicator-delay: var(--media-loading-indicator-transition-delay, ${xs}ms);
      }

      #status {
        color: rgba(0,0,0,0);
        width: 0px;
        height: 0px;
      }

      :host slot[name=icon] > *,
      :host ::slotted([slot=icon]) {
        opacity: var(--media-loading-indicator-opacity, 0);
        transition: opacity 0.15s;
      }

      :host([${b.MEDIA_LOADING}]:not([${b.MEDIA_PAUSED}])) slot[name=icon] > *,
      :host([${b.MEDIA_LOADING}]:not([${b.MEDIA_PAUSED}])) ::slotted([slot=icon]) {
        opacity: var(--media-loading-indicator-opacity, 1);
        transition: opacity 0.15s var(--_loading-indicator-delay);
      }

      :host #status {
        visibility: var(--media-loading-indicator-opacity, hidden);
        transition: visibility 0.15s;
      }

      :host([${b.MEDIA_LOADING}]:not([${b.MEDIA_PAUSED}])) #status {
        visibility: var(--media-loading-indicator-opacity, visible);
        transition: visibility 0.15s var(--_loading-indicator-delay);
      }

      svg, img, ::slotted(svg), ::slotted(img) {
        width: var(--media-loading-indicator-icon-width);
        height: var(--media-loading-indicator-icon-height, 100px);
        fill: var(--media-icon-color, var(--media-primary-color, rgb(238 238 238)));
        vertical-align: middle;
      }
    </style>

    <slot name="icon">${Ss}</slot>
    <div id="status" role="status" aria-live="polite">${E(`media loading`)}</div>
  `}var ws=class extends O.HTMLElement{constructor(){if(super(),gs(this,vs,void 0),gs(this,ys,xs),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}}static get observedAttributes(){return[y.MEDIA_CONTROLLER,b.MEDIA_PAUSED,b.MEDIA_LOADING,bs.LOADING_DELAY]}attributeChangedCallback(e,t,n){var r,i,a,o;e===bs.LOADING_DELAY&&t!==n?this.loadingDelay=Number(n):e===y.MEDIA_CONTROLLER&&(t&&((i=(r=hs(this,vs))?.unassociateElement)==null||i.call(r,this),_s(this,vs,null)),n&&this.isConnected&&(_s(this,vs,this.getRootNode()?.getElementById(n)),(o=(a=hs(this,vs))?.associateElement)==null||o.call(a,this)))}connectedCallback(){var e,t;let n=this.getAttribute(y.MEDIA_CONTROLLER);n&&(_s(this,vs,this.getRootNode()?.getElementById(n)),(t=(e=hs(this,vs))?.associateElement)==null||t.call(e,this))}disconnectedCallback(){var e,t;(t=(e=hs(this,vs))?.unassociateElement)==null||t.call(e,this),_s(this,vs,null)}get loadingDelay(){return hs(this,ys)}set loadingDelay(e){_s(this,ys,e);let{style:t}=A(this.shadowRoot,`:host`);t.setProperty(`--_loading-indicator-delay`,`var(--media-loading-indicator-transition-delay, ${e}ms)`)}get mediaPaused(){return N(this,b.MEDIA_PAUSED)}set mediaPaused(e){P(this,b.MEDIA_PAUSED,e)}get mediaLoading(){return N(this,b.MEDIA_LOADING)}set mediaLoading(e){P(this,b.MEDIA_LOADING,e)}get mediaController(){return F(this,y.MEDIA_CONTROLLER)}set mediaController(e){I(this,y.MEDIA_CONTROLLER,e)}get noAutohide(){return N(this,bs.NO_AUTOHIDE)}set noAutohide(e){P(this,bs.NO_AUTOHIDE,e)}};vs=new WeakMap,ys=new WeakMap,ws.shadowRootOptions={mode:`open`},ws.getTemplateHTML=Cs,O.customElements.get(`media-loading-indicator`)||O.customElements.define(`media-loading-indicator`,ws);var Ts=ws,Es=`<svg aria-hidden="true" viewBox="0 0 24 24">
  <path d="M16.5 12A4.5 4.5 0 0 0 14 8v2.18l2.45 2.45a4.22 4.22 0 0 0 .05-.63Zm2.5 0a6.84 6.84 0 0 1-.54 2.64L20 16.15A8.8 8.8 0 0 0 21 12a9 9 0 0 0-7-8.77v2.06A7 7 0 0 1 19 12ZM4.27 3 3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25A6.92 6.92 0 0 1 14 18.7v2.06A9 9 0 0 0 17.69 19l2 2.05L21 19.73l-9-9L4.27 3ZM12 4 9.91 6.09 12 8.18V4Z"/>
</svg>`,Ds=`<svg aria-hidden="true" viewBox="0 0 24 24">
  <path d="M3 9v6h4l5 5V4L7 9H3Zm13.5 3A4.5 4.5 0 0 0 14 8v8a4.47 4.47 0 0 0 2.5-4Z"/>
</svg>`,Os=`<svg aria-hidden="true" viewBox="0 0 24 24">
  <path d="M3 9v6h4l5 5V4L7 9H3Zm13.5 3A4.5 4.5 0 0 0 14 8v8a4.47 4.47 0 0 0 2.5-4ZM14 3.23v2.06a7 7 0 0 1 0 13.42v2.06a9 9 0 0 0 0-17.54Z"/>
</svg>`;function ks(e){return`
    <style>
      :host(:not([${b.MEDIA_VOLUME_LEVEL}])) slot[name=icon] slot:not([name=high]),
      :host([${b.MEDIA_VOLUME_LEVEL}=high]) slot[name=icon] slot:not([name=high]) {
        display: none !important;
      }

      :host([${b.MEDIA_VOLUME_LEVEL}=off]) slot[name=icon] slot:not([name=off]) {
        display: none !important;
      }

      :host([${b.MEDIA_VOLUME_LEVEL}=low]) slot[name=icon] slot:not([name=low]) {
        display: none !important;
      }

      :host([${b.MEDIA_VOLUME_LEVEL}=medium]) slot[name=icon] slot:not([name=medium]) {
        display: none !important;
      }

      :host(:not([${b.MEDIA_VOLUME_LEVEL}=off])) slot[name=tooltip-unmute],
      :host([${b.MEDIA_VOLUME_LEVEL}=off]) slot[name=tooltip-mute] {
        display: none;
      }
    </style>

    <slot name="icon">
      <slot name="off">${Es}</slot>
      <slot name="low">${Ds}</slot>
      <slot name="medium">${Ds}</slot>
      <slot name="high">${Os}</slot>
    </slot>
  `}function As(){return`
    <slot name="tooltip-mute">${E(`Mute`)}</slot>
    <slot name="tooltip-unmute">${E(`Unmute`)}</slot>
  `}var js=e=>{let t=e.mediaVolumeLevel===`off`?E(`unmute`):E(`mute`);e.setAttribute(`aria-label`,t)},Ms=class extends K{static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_VOLUME_LEVEL]}connectedCallback(){super.connectedCallback(),js(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===b.MEDIA_VOLUME_LEVEL&&js(this)}get mediaVolumeLevel(){return F(this,b.MEDIA_VOLUME_LEVEL)}set mediaVolumeLevel(e){I(this,b.MEDIA_VOLUME_LEVEL,e)}handleClick(){let e=this.mediaVolumeLevel===`off`?v.MEDIA_UNMUTE_REQUEST:v.MEDIA_MUTE_REQUEST;this.dispatchEvent(new O.CustomEvent(e,{composed:!0,bubbles:!0}))}};Ms.getSlotTemplateHTML=ks,Ms.getTooltipContentHTML=As,O.customElements.get(`media-mute-button`)||O.customElements.define(`media-mute-button`,Ms);var Ns=Ms,Ps=`<svg aria-hidden="true" viewBox="0 0 28 24">
  <path d="M24 3H4a1 1 0 0 0-1 1v16a1 1 0 0 0 1 1h20a1 1 0 0 0 1-1V4a1 1 0 0 0-1-1Zm-1 16H5V5h18v14Zm-3-8h-7v5h7v-5Z"/>
</svg>`;function Fs(e){return`
    <style>
      :host([${b.MEDIA_IS_PIP}]) slot[name=icon] slot:not([name=exit]) {
        display: none !important;
      }

      :host(:not([${b.MEDIA_IS_PIP}])) slot[name=icon] slot:not([name=enter]) {
        display: none !important;
      }

      :host([${b.MEDIA_IS_PIP}]) slot[name=tooltip-enter],
      :host(:not([${b.MEDIA_IS_PIP}])) slot[name=tooltip-exit] {
        display: none;
      }
    </style>

    <slot name="icon">
      <slot name="enter">${Ps}</slot>
      <slot name="exit">${Ps}</slot>
    </slot>
  `}function Is(){return`
    <slot name="tooltip-enter">${E(`Enter picture in picture mode`)}</slot>
    <slot name="tooltip-exit">${E(`Exit picture in picture mode`)}</slot>
  `}var Ls=e=>{let t=e.mediaIsPip?E(`exit picture in picture mode`):E(`enter picture in picture mode`);e.setAttribute(`aria-label`,t)},Rs=class extends K{static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_IS_PIP,b.MEDIA_PIP_UNAVAILABLE]}connectedCallback(){super.connectedCallback(),Ls(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===b.MEDIA_IS_PIP&&Ls(this)}get mediaPipUnavailable(){return F(this,b.MEDIA_PIP_UNAVAILABLE)}set mediaPipUnavailable(e){I(this,b.MEDIA_PIP_UNAVAILABLE,e)}get mediaIsPip(){return N(this,b.MEDIA_IS_PIP)}set mediaIsPip(e){P(this,b.MEDIA_IS_PIP,e)}handleClick(){let e=this.mediaIsPip?v.MEDIA_EXIT_PIP_REQUEST:v.MEDIA_ENTER_PIP_REQUEST;this.dispatchEvent(new O.CustomEvent(e,{composed:!0,bubbles:!0}))}};Rs.getSlotTemplateHTML=Fs,Rs.getTooltipContentHTML=Is,O.customElements.get(`media-pip-button`)||O.customElements.define(`media-pip-button`,Rs);var zs=Rs,Bs=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Vs=(e,t,n)=>(Bs(e,t,`read from private field`),n?n.call(e):t.get(e)),Hs=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Us,Ws={RATES:`rates`},Gs=[1,1.2,1.5,1.7,2];function Ks(e){return Math.round(e*100)/100}function qs(e){return`
    <style>
      :host {
        min-width: 5ch;
        padding: var(--media-button-padding, var(--media-control-padding, 10px 5px));
      }
    </style>
    <slot name="icon">${e.mediaplaybackrate?Ks(+e.mediaplaybackrate):1}x</slot>
  `}function Js(){return E(`Playback rate`)}var Ys=class extends K{constructor(){super(),Hs(this,Us,new sr(this,Ws.RATES,{defaultValue:Gs})),this.container=this.shadowRoot.querySelector(`slot[name="icon"]`),this.container.innerHTML=`${Ks(this.mediaPlaybackRate??1)}x`}static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_PLAYBACK_RATE,Ws.RATES]}attributeChangedCallback(e,t,n){if(super.attributeChangedCallback(e,t,n),e===Ws.RATES&&(Vs(this,Us).value=n),e===b.MEDIA_PLAYBACK_RATE){let e=n?+n:NaN,t=Ks(Number.isNaN(e)?1:e);this.container.innerHTML=`${t}x`,this.setAttribute(`aria-label`,E(`Playback rate {playbackRate}`,{playbackRate:t}))}}get rates(){return Vs(this,Us)}set rates(e){e?Array.isArray(e)?Vs(this,Us).value=e.join(` `):typeof e==`string`&&(Vs(this,Us).value=e):Vs(this,Us).value=``}get mediaPlaybackRate(){return j(this,b.MEDIA_PLAYBACK_RATE,1)}set mediaPlaybackRate(e){M(this,b.MEDIA_PLAYBACK_RATE,e)}handleClick(){let e=Array.from(Vs(this,Us).values(),e=>+e).sort((e,t)=>e-t),t=e.find(e=>e>this.mediaPlaybackRate)??e[0]??1,n=new O.CustomEvent(v.MEDIA_PLAYBACK_RATE_REQUEST,{composed:!0,bubbles:!0,detail:t});this.dispatchEvent(n)}};Us=new WeakMap,Ys.getSlotTemplateHTML=qs,Ys.getTooltipContentHTML=Js,O.customElements.get(`media-playback-rate-button`)||O.customElements.define(`media-playback-rate-button`,Ys);var Xs=Ys,Zs=`<svg aria-hidden="true" viewBox="0 0 24 24">
  <path d="m6 21 15-9L6 3v18Z"/>
</svg>`,Qs=`<svg aria-hidden="true" viewBox="0 0 24 24">
  <path d="M6 20h4V4H6v16Zm8-16v16h4V4h-4Z"/>
</svg>`;function $s(e){return`
    <style>
      :host([${b.MEDIA_PAUSED}]) slot[name=pause],
      :host(:not([${b.MEDIA_PAUSED}])) slot[name=play] {
        display: none !important;
      }

      :host([${b.MEDIA_PAUSED}]) slot[name=tooltip-pause],
      :host(:not([${b.MEDIA_PAUSED}])) slot[name=tooltip-play] {
        display: none;
      }
    </style>

    <slot name="icon">
      <slot name="play">${Zs}</slot>
      <slot name="pause">${Qs}</slot>
    </slot>
  `}function ec(){return`
    <slot name="tooltip-play">${E(`Play`)}</slot>
    <slot name="tooltip-pause">${E(`Pause`)}</slot>
  `}var tc=e=>{let t=e.mediaPaused?E(`play`):E(`pause`);e.setAttribute(`aria-label`,t)},nc=class extends K{static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_PAUSED,b.MEDIA_ENDED]}connectedCallback(){super.connectedCallback(),tc(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),(e===b.MEDIA_PAUSED||e===b.MEDIA_LANG)&&tc(this)}get mediaPaused(){return N(this,b.MEDIA_PAUSED)}set mediaPaused(e){P(this,b.MEDIA_PAUSED,e)}handleClick(){let e=this.mediaPaused?v.MEDIA_PLAY_REQUEST:v.MEDIA_PAUSE_REQUEST;this.dispatchEvent(new O.CustomEvent(e,{composed:!0,bubbles:!0}))}};nc.getSlotTemplateHTML=$s,nc.getTooltipContentHTML=ec,O.customElements.get(`media-play-button`)||O.customElements.define(`media-play-button`,nc);var rc=nc,ic={PLACEHOLDER_SRC:`placeholdersrc`,SRC:`src`};function ac(e){return`
    <style>
      :host {
        pointer-events: none;
        display: var(--media-poster-image-display, inline-block);
        box-sizing: border-box;
      }

      img {
        max-width: 100%;
        max-height: 100%;
        min-width: 100%;
        min-height: 100%;
        background-repeat: no-repeat;
        background-position: var(--media-poster-image-background-position, var(--media-object-position, center));
        background-size: var(--media-poster-image-background-size, var(--media-object-fit, contain));
        object-fit: var(--media-object-fit, contain);
        object-position: var(--media-object-position, center);
      }
    </style>

    <img part="poster img" aria-hidden="true" id="image"/>
  `}var oc=e=>{e.style.removeProperty(`background-image`)},sc=(e,t)=>{e.style[`background-image`]=`url('${t}')`},cc=class extends O.HTMLElement{static get observedAttributes(){return[ic.PLACEHOLDER_SRC,ic.SRC]}constructor(){if(super(),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}this.image=this.shadowRoot.querySelector(`#image`)}attributeChangedCallback(e,t,n){e===ic.SRC&&(n==null?this.image.removeAttribute(ic.SRC):this.image.setAttribute(ic.SRC,n)),e===ic.PLACEHOLDER_SRC&&(n==null?oc(this.image):sc(this.image,n))}get placeholderSrc(){return F(this,ic.PLACEHOLDER_SRC)}set placeholderSrc(e){I(this,ic.SRC,e)}get src(){return F(this,ic.SRC)}set src(e){I(this,ic.SRC,e)}};cc.shadowRootOptions={mode:`open`},cc.getTemplateHTML=ac,O.customElements.get(`media-poster-image`)||O.customElements.define(`media-poster-image`,cc);var lc=cc,uc=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},dc=(e,t,n)=>(uc(e,t,`read from private field`),n?n.call(e):t.get(e)),fc=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},pc=(e,t,n,r)=>(uc(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),mc,hc=class extends Co{constructor(){super(),fc(this,mc,void 0),pc(this,mc,this.shadowRoot.querySelector(`slot`))}static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_PREVIEW_CHAPTER,b.MEDIA_LANG]}attributeChangedCallback(e,t,n){if(super.attributeChangedCallback(e,t,n),(e===b.MEDIA_PREVIEW_CHAPTER||e===b.MEDIA_LANG)&&n!==t&&n!=null){if(dc(this,mc).textContent=n,n!==``){let e=E(`chapter: {chapterName}`,{chapterName:n});this.setAttribute(`aria-valuetext`,e)}else this.removeAttribute(`aria-valuetext`)}}get mediaPreviewChapter(){return F(this,b.MEDIA_PREVIEW_CHAPTER)}set mediaPreviewChapter(e){I(this,b.MEDIA_PREVIEW_CHAPTER,e)}};mc=new WeakMap,O.customElements.get(`media-preview-chapter-display`)||O.customElements.define(`media-preview-chapter-display`,hc);var gc=hc,_c=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},vc=(e,t,n)=>(_c(e,t,`read from private field`),n?n.call(e):t.get(e)),yc=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},bc=(e,t,n,r)=>(_c(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),xc;function Sc(e){return`
    <style>
      :host {
        box-sizing: border-box;
        display: var(--media-control-display, var(--media-preview-thumbnail-display, inline-block));
        overflow: hidden;
      }

      img {
        display: none;
        position: relative;
      }
    </style>
    <img crossorigin loading="eager" decoding="async">
  `}var Cc=class extends O.HTMLElement{constructor(){if(super(),yc(this,xc,void 0),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}}static get observedAttributes(){return[y.MEDIA_CONTROLLER,b.MEDIA_PREVIEW_IMAGE,b.MEDIA_PREVIEW_COORDS]}connectedCallback(){var e,t;let n=this.getAttribute(y.MEDIA_CONTROLLER);n&&(bc(this,xc,this.getRootNode()?.getElementById(n)),(t=(e=vc(this,xc))?.associateElement)==null||t.call(e,this))}disconnectedCallback(){var e,t;(t=(e=vc(this,xc))?.unassociateElement)==null||t.call(e,this),bc(this,xc,null)}attributeChangedCallback(e,t,n){var r,i,a,o;[b.MEDIA_PREVIEW_IMAGE,b.MEDIA_PREVIEW_COORDS].includes(e)&&this.update(),e===y.MEDIA_CONTROLLER&&(t&&((i=(r=vc(this,xc))?.unassociateElement)==null||i.call(r,this),bc(this,xc,null)),n&&this.isConnected&&(bc(this,xc,this.getRootNode()?.getElementById(n)),(o=(a=vc(this,xc))?.associateElement)==null||o.call(a,this)))}get mediaPreviewImage(){return F(this,b.MEDIA_PREVIEW_IMAGE)}set mediaPreviewImage(e){I(this,b.MEDIA_PREVIEW_IMAGE,e)}get mediaPreviewCoords(){let e=this.getAttribute(b.MEDIA_PREVIEW_COORDS);if(e)return e.split(/\s+/).map(e=>+e)}set mediaPreviewCoords(e){e?this.setAttribute(b.MEDIA_PREVIEW_COORDS,e.join(` `)):this.removeAttribute(b.MEDIA_PREVIEW_COORDS)}update(){let e=this.mediaPreviewCoords,t=this.mediaPreviewImage;if(!(e&&t))return;let[n,r,i,a]=e,o=t.split(`#`)[0],s=getComputedStyle(this),{maxWidth:c,maxHeight:l,minWidth:u,minHeight:d}=s,f=s.getPropertyValue(`--media-preview-thumbnail-object-fit`).trim()||`contain`,p,m;if(f===`fill`){let e=parseInt(c)/i,t=parseInt(l)/a,n=parseInt(u)/i,r=parseInt(d)/a;p=e<1?e:Math.max(e,n),m=t<1?t:Math.max(t,r)}else{let e=Math.min(parseInt(c)/i,parseInt(l)/a),t=Math.max(parseInt(u)/i,parseInt(d)/a),n=e<1?e:t>1?t:1;p=n,m=n}let{style:h}=A(this.shadowRoot,`:host`),ee=A(this.shadowRoot,`img`).style,te=this.shadowRoot.querySelector(`img`),ne=Math.min(p,m)<1?`min`:`max`;h.setProperty(`${ne}-width`,`initial`,`important`),h.setProperty(`${ne}-height`,`initial`,`important`),h.width=`${i*p}px`,h.height=`${a*m}px`;let re=()=>{ee.width=`${this.imgWidth*p}px`,ee.height=`${this.imgHeight*m}px`,ee.display=`block`};te.src!==o&&(te.onload=()=>{this.imgWidth=te.naturalWidth,this.imgHeight=te.naturalHeight,re(),te.onload=null},te.src=o,re()),re(),ee.transform=`translate(-${n*p}px, -${r*m}px)`}};xc=new WeakMap,Cc.shadowRootOptions={mode:`open`},Cc.getTemplateHTML=Sc,O.customElements.get(`media-preview-thumbnail`)||O.customElements.define(`media-preview-thumbnail`,Cc);var wc=Cc,Tc=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Ec=(e,t,n)=>(Tc(e,t,`read from private field`),n?n.call(e):t.get(e)),Dc=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Oc=(e,t,n,r)=>(Tc(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),kc,Ac=class extends Co{constructor(){super(),Dc(this,kc,void 0),Oc(this,kc,this.shadowRoot.querySelector(`slot`)),Ec(this,kc).textContent=D(0)}static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_PREVIEW_TIME]}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===b.MEDIA_PREVIEW_TIME&&n!=null&&(Ec(this,kc).textContent=D(parseFloat(n)))}get mediaPreviewTime(){return j(this,b.MEDIA_PREVIEW_TIME)}set mediaPreviewTime(e){M(this,b.MEDIA_PREVIEW_TIME,e)}};kc=new WeakMap,O.customElements.get(`media-preview-time-display`)||O.customElements.define(`media-preview-time-display`,Ac);var jc=Ac,Mc={SEEK_OFFSET:`seekoffset`},Nc=30,Pc=e=>`
  <svg aria-hidden="true" viewBox="0 0 20 24">
    <defs>
      <style>.text{font-size:8px;font-family:Arial-BoldMT, Arial;font-weight:700;}</style>
    </defs>
    <text class="text value" transform="translate(2.18 19.87)">${e}</text>
    <path d="M10 6V3L4.37 7 10 10.94V8a5.54 5.54 0 0 1 1.9 10.48v2.12A7.5 7.5 0 0 0 10 6Z"/>
  </svg>`;function Fc(e,t){return`
    <slot name="icon">${Pc(t.seekOffset)}</slot>
  `}var Ic=(e,t)=>{e.setAttribute(`aria-label`,E(`seek back {seekOffset} seconds`,{seekOffset:t}))};function Lc(){return E(`Seek backward`)}var Rc=0,zc=class extends K{static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_CURRENT_TIME,Mc.SEEK_OFFSET]}connectedCallback(){super.connectedCallback(),this.seekOffset=j(this,Mc.SEEK_OFFSET,Nc)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),Ic(this,this.seekOffset),e===Mc.SEEK_OFFSET&&(this.seekOffset=j(this,Mc.SEEK_OFFSET,Nc))}get seekOffset(){return j(this,Mc.SEEK_OFFSET,Nc)}set seekOffset(e){M(this,Mc.SEEK_OFFSET,e),this.setAttribute(`aria-label`,E(`seek back {seekOffset} seconds`,{seekOffset:this.seekOffset})),gt(vt(this,`icon`),this.seekOffset)}get mediaCurrentTime(){return j(this,b.MEDIA_CURRENT_TIME,Rc)}set mediaCurrentTime(e){M(this,b.MEDIA_CURRENT_TIME,e)}handleClick(){let e=Math.max(this.mediaCurrentTime-this.seekOffset,0),t=new O.CustomEvent(v.MEDIA_SEEK_REQUEST,{composed:!0,bubbles:!0,detail:e});this.dispatchEvent(t)}};zc.getSlotTemplateHTML=Fc,zc.getTooltipContentHTML=Lc,O.customElements.get(`media-seek-backward-button`)||O.customElements.define(`media-seek-backward-button`,zc);var Bc=zc,Vc={SEEK_OFFSET:`seekoffset`},Hc=30,Uc=e=>`
  <svg aria-hidden="true" viewBox="0 0 20 24">
    <defs>
      <style>.text{font-size:8px;font-family:Arial-BoldMT, Arial;font-weight:700;}</style>
    </defs>
    <text class="text value" transform="translate(8.9 19.87)">${e}</text>
    <path d="M10 6V3l5.61 4L10 10.94V8a5.54 5.54 0 0 0-1.9 10.48v2.12A7.5 7.5 0 0 1 10 6Z"/>
  </svg>`;function Wc(e,t){return`
    <slot name="icon">${Uc(t.seekOffset)}</slot>
  `}var Gc=(e,t)=>{e.setAttribute(`aria-label`,E(`seek forward {seekOffset} seconds`,{seekOffset:t}))};function Kc(){return E(`Seek forward`)}var qc=0,Jc=class extends K{static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_CURRENT_TIME,Vc.SEEK_OFFSET]}connectedCallback(){super.connectedCallback(),this.seekOffset=j(this,Vc.SEEK_OFFSET,Hc)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),Gc(this,this.seekOffset),e===Vc.SEEK_OFFSET&&(this.seekOffset=j(this,Vc.SEEK_OFFSET,Hc))}get seekOffset(){return j(this,Vc.SEEK_OFFSET,Hc)}set seekOffset(e){M(this,Vc.SEEK_OFFSET,e),this.setAttribute(`aria-label`,E(`seek forward {seekOffset} seconds`,{seekOffset:this.seekOffset})),gt(vt(this,`icon`),this.seekOffset)}get mediaCurrentTime(){return j(this,b.MEDIA_CURRENT_TIME,qc)}set mediaCurrentTime(e){M(this,b.MEDIA_CURRENT_TIME,e)}handleClick(){let e=this.mediaCurrentTime+this.seekOffset,t=new O.CustomEvent(v.MEDIA_SEEK_REQUEST,{composed:!0,bubbles:!0,detail:e});this.dispatchEvent(t)}};Jc.getSlotTemplateHTML=Wc,Jc.getTooltipContentHTML=Kc,O.customElements.get(`media-seek-forward-button`)||O.customElements.define(`media-seek-forward-button`,Jc);var Yc=Jc,Xc=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Zc=(e,t,n)=>(Xc(e,t,`read from private field`),n?n.call(e):t.get(e)),Qc=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},$c=(e,t,n,r)=>(Xc(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),el=(e,t,n)=>(Xc(e,t,`access private method`),n),tl,nl,rl,il,al,ol,sl,cl,ll,ul,dl,fl={REMAINING:`remaining`,SHOW_DURATION:`showduration`,NO_TOGGLE:`notoggle`},pl=[...Object.values(fl),b.MEDIA_CURRENT_TIME,b.MEDIA_DURATION,b.MEDIA_SEEKABLE],ml=[`Enter`,` `],hl=`&nbsp;/&nbsp;`,gl=(e,{timesSep:t=hl}={})=>{let n=e.mediaCurrentTime??0,[,r]=e.mediaSeekable??[],i=0;Number.isFinite(e.mediaDuration)?i=e.mediaDuration:Number.isFinite(r)&&(i=r);let a=e.remaining?D(0-(i-n)):D(n);return e.showDuration?`${a}${t}${D(i)}`:a},_l=e=>{let t=e.mediaCurrentTime,[,n]=e.mediaSeekable??[],r=null;if(Number.isFinite(e.mediaDuration)?r=e.mediaDuration:Number.isFinite(n)&&(r=n),t==null||r===null){e.setAttribute(`aria-description`,E(`video not loaded, unknown time.`));return}let i=e.remaining?$e(0-(r-t)):$e(t);if(!e.showDuration){e.setAttribute(`aria-description`,i);return}let a=E(`{currentTime} of {totalTime}`,{currentTime:i,totalTime:$e(r)});e.setAttribute(`aria-description`,a)};function vl(e,t){return`
    <slot>${gl(t)}</slot>
  `}var yl=e=>{e.setAttribute(`aria-label`,E(`playback time`))},bl=class extends Co{constructor(){super(),Qc(this,il),Qc(this,ol),Qc(this,cl),Qc(this,ul),Qc(this,tl,void 0),Qc(this,nl,null),Qc(this,rl,e=>{let{metaKey:t,altKey:n,key:r}=e;t||n||!ml.includes(r)?this.removeEventListener(`keyup`,Zc(this,nl)):this.addEventListener(`keyup`,Zc(this,nl))}),$c(this,tl,this.shadowRoot.querySelector(`slot`)),Zc(this,tl).innerHTML=`${gl(this)}`}static get observedAttributes(){return[...super.observedAttributes,...pl,`disabled`]}connectedCallback(){let{style:e}=A(this.shadowRoot,`:host(:hover:not([notoggle]))`);e.setProperty(`cursor`,`var(--media-cursor, pointer)`),e.setProperty(`background`,`var(--media-control-hover-background, rgba(50 50 70 / .7))`),this.setAttribute(`aria-label`,E(`playback time`)),el(this,cl,ll).call(this),super.connectedCallback()}toggleTimeDisplay(){this.noToggle||(this.hasAttribute(`remaining`)?this.removeAttribute(`remaining`):this.setAttribute(`remaining`,``))}disconnectedCallback(){this.disable(),el(this,ol,sl).call(this),super.disconnectedCallback()}attributeChangedCallback(e,t,n){yl(this),pl.includes(e)?this.update():e===`disabled`&&n!==t?n==null?el(this,cl,ll).call(this):el(this,ul,dl).call(this):e===fl.NO_TOGGLE&&n!==t&&(this.noToggle?el(this,ul,dl).call(this):el(this,cl,ll).call(this)),super.attributeChangedCallback(e,t,n)}enable(){this.noToggle||(this.tabIndex=0)}disable(){this.tabIndex=-1}get remaining(){return N(this,fl.REMAINING)}set remaining(e){P(this,fl.REMAINING,e)}get showDuration(){return N(this,fl.SHOW_DURATION)}set showDuration(e){P(this,fl.SHOW_DURATION,e)}get noToggle(){return N(this,fl.NO_TOGGLE)}set noToggle(e){P(this,fl.NO_TOGGLE,e)}get mediaDuration(){return j(this,b.MEDIA_DURATION)}set mediaDuration(e){M(this,b.MEDIA_DURATION,e)}get mediaCurrentTime(){return j(this,b.MEDIA_CURRENT_TIME)}set mediaCurrentTime(e){M(this,b.MEDIA_CURRENT_TIME,e)}get mediaSeekable(){let e=this.getAttribute(b.MEDIA_SEEKABLE);if(e)return e.split(`:`).map(e=>+e)}set mediaSeekable(e){e==null?this.removeAttribute(b.MEDIA_SEEKABLE):this.setAttribute(b.MEDIA_SEEKABLE,e.join(`:`))}update(){let e=gl(this);_l(this),e!==Zc(this,tl).innerHTML&&(Zc(this,tl).innerHTML=e)}};tl=new WeakMap,nl=new WeakMap,rl=new WeakMap,il=new WeakSet,al=function(){Zc(this,nl)||($c(this,nl,e=>{let{key:t}=e;ml.includes(t)?this.toggleTimeDisplay():this.removeEventListener(`keyup`,Zc(this,nl))}),this.addEventListener(`keydown`,Zc(this,rl)),this.addEventListener(`click`,this.toggleTimeDisplay))},ol=new WeakSet,sl=function(){Zc(this,nl)&&(this.removeEventListener(`keyup`,Zc(this,nl)),this.removeEventListener(`keydown`,Zc(this,rl)),this.removeEventListener(`click`,this.toggleTimeDisplay),$c(this,nl,null))},cl=new WeakSet,ll=function(){!this.noToggle&&!this.hasAttribute(`disabled`)&&(this.setAttribute(`role`,`button`),this.enable(),el(this,il,al).call(this))},ul=new WeakSet,dl=function(){this.removeAttribute(`role`),this.disable(),el(this,ol,sl).call(this)},bl.getSlotTemplateHTML=vl,O.customElements.get(`media-time-display`)||O.customElements.define(`media-time-display`,bl);var xl=bl,Sl=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Y=(e,t,n)=>(Sl(e,t,`read from private field`),n?n.call(e):t.get(e)),Cl=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},wl=(e,t,n,r)=>(Sl(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),Tl=(e,t,n,r)=>({set _(r){wl(e,t,r,n)},get _(){return Y(e,t,r)}}),El,Dl,Ol,kl,Al,jl,Ml,Nl,Pl,Fl,Il=class{constructor(e,t,n){Cl(this,El,void 0),Cl(this,Dl,void 0),Cl(this,Ol,void 0),Cl(this,kl,void 0),Cl(this,Al,void 0),Cl(this,jl,void 0),Cl(this,Ml,void 0),Cl(this,Nl,void 0),Cl(this,Pl,0),Cl(this,Fl,(e=performance.now())=>{wl(this,Pl,requestAnimationFrame(Y(this,Fl))),wl(this,kl,performance.now()-Y(this,Ol));let t=1e3/this.fps;if(Y(this,kl)>t){wl(this,Ol,e-Y(this,kl)%t);let n=1e3/((e-Y(this,Dl))/++Tl(this,Al)._),r=(e-Y(this,jl))/1e3/this.duration,i=Y(this,Ml)+r*this.playbackRate;i-Y(this,El).valueAsNumber>0?wl(this,Nl,this.playbackRate/this.duration/n):(wl(this,Nl,.995*Y(this,Nl)),i=Y(this,El).valueAsNumber+Y(this,Nl)),this.callback(i)}}),wl(this,El,e),this.callback=t,this.fps=n}start(){Y(this,Pl)===0&&(wl(this,Ol,performance.now()),wl(this,Dl,Y(this,Ol)),wl(this,Al,0),Y(this,Fl).call(this))}stop(){Y(this,Pl)!==0&&(cancelAnimationFrame(Y(this,Pl)),wl(this,Pl,0))}update({start:e,duration:t,playbackRate:n}){let r=e-Y(this,El).valueAsNumber,i=Math.abs(t-this.duration);(r>0||r<-.5/t||i>=.5)&&this.callback(e),wl(this,Ml,e),wl(this,jl,performance.now()),this.duration=t,this.playbackRate=n}};El=new WeakMap,Dl=new WeakMap,Ol=new WeakMap,kl=new WeakMap,Al=new WeakMap,jl=new WeakMap,Ml=new WeakMap,Nl=new WeakMap,Pl=new WeakMap,Fl=new WeakMap;var Ll=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},X=(e,t,n)=>(Ll(e,t,`read from private field`),n?n.call(e):t.get(e)),Z=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Rl=(e,t,n,r)=>(Ll(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),zl=(e,t,n)=>(Ll(e,t,`access private method`),n),Bl,Vl,Hl,Ul,Wl,Gl,Kl,ql,Jl,Yl,Xl,Zl,Ql,$l,eu,tu,nu,ru,iu,au,ou,su,cu,lu,uu,du,fu=e=>{let t=e.range,n=$e(+hu(e)),r=$e(+e.mediaSeekableEnd),i=n&&r?E(`{currentTime} of {totalTime}`,{currentTime:n,totalTime:r}):E(`video not loaded, unknown time.`);t.setAttribute(`aria-valuetext`,i)};function pu(e){return`
    <style>
      :host {
        --media-box-border-radius: 4px;
        --media-box-padding-left: 10px;
        --media-box-padding-right: 10px;
        --media-preview-border-radius: var(--media-box-border-radius);
        --media-box-arrow-offset: var(--media-box-border-radius);
        --_control-background: var(--media-control-background, var(--media-secondary-color, rgb(20 20 30 / .7)));
        --_preview-background: var(--media-preview-background, var(--_control-background));

        
        contain: layout;
      }

      #buffered {
        background: var(--media-time-range-buffered-color, rgb(255 255 255 / .4));
        position: absolute;
        height: 100%;
        will-change: width;
      }

      #preview-rail,
      #current-rail {
        width: 100%;
        position: absolute;
        left: 0;
        bottom: 100%;
        pointer-events: none;
        will-change: transform;
      }

      [part~="box"] {
        width: min-content;
        
        position: absolute;
        bottom: 100%;
        flex-direction: column;
        align-items: center;
        transform: translateX(-50%);
      }

      [part~="current-box"] {
        display: var(--media-current-box-display, var(--media-box-display, flex));
        margin: var(--media-current-box-margin, var(--media-box-margin, 0 0 5px));
        visibility: hidden;
      }

      [part~="preview-box"] {
        display: var(--media-preview-box-display, var(--media-box-display, flex));
        margin: var(--media-preview-box-margin, var(--media-box-margin, 0 0 5px));
        transition-property: var(--media-preview-transition-property, visibility, opacity);
        transition-duration: var(--media-preview-transition-duration-out, .25s);
        transition-delay: var(--media-preview-transition-delay-out, 0s);
        visibility: hidden;
        opacity: 0;
      }

      :host(:is([${b.MEDIA_PREVIEW_IMAGE}], [${b.MEDIA_PREVIEW_TIME}])[dragging]) [part~="preview-box"] {
        transition-duration: var(--media-preview-transition-duration-in, .5s);
        transition-delay: var(--media-preview-transition-delay-in, .25s);
        visibility: visible;
        opacity: 1;
      }

      @media (hover: hover) {
        :host(:is([${b.MEDIA_PREVIEW_IMAGE}], [${b.MEDIA_PREVIEW_TIME}]):hover) [part~="preview-box"] {
          transition-duration: var(--media-preview-transition-duration-in, .5s);
          transition-delay: var(--media-preview-transition-delay-in, .25s);
          visibility: visible;
          opacity: 1;
        }
      }

      media-preview-thumbnail,
      ::slotted(media-preview-thumbnail) {
        visibility: hidden;
        
        transition: visibility 0s .25s;
        transition-delay: calc(var(--media-preview-transition-delay-out, 0s) + var(--media-preview-transition-duration-out, .25s));
        background: var(--media-preview-thumbnail-background, var(--_preview-background));
        box-shadow: var(--media-preview-thumbnail-box-shadow, 0 0 4px rgb(0 0 0 / .2));
        max-width: var(--media-preview-thumbnail-max-width, 180px);
        max-height: var(--media-preview-thumbnail-max-height, 160px);
        min-width: var(--media-preview-thumbnail-min-width, 120px);
        min-height: var(--media-preview-thumbnail-min-height, 80px);
        border: var(--media-preview-thumbnail-border);
        border-radius: var(--media-preview-thumbnail-border-radius,
          var(--media-preview-border-radius) var(--media-preview-border-radius) 0 0);
      }

      :host([${b.MEDIA_PREVIEW_IMAGE}][dragging]) media-preview-thumbnail,
      :host([${b.MEDIA_PREVIEW_IMAGE}][dragging]) ::slotted(media-preview-thumbnail) {
        transition-delay: var(--media-preview-transition-delay-in, .25s);
        visibility: visible;
      }

      @media (hover: hover) {
        :host([${b.MEDIA_PREVIEW_IMAGE}]:hover) media-preview-thumbnail,
        :host([${b.MEDIA_PREVIEW_IMAGE}]:hover) ::slotted(media-preview-thumbnail) {
          transition-delay: var(--media-preview-transition-delay-in, .25s);
          visibility: visible;
        }

        :host([${b.MEDIA_PREVIEW_TIME}]:hover) {
          --media-time-range-hover-display: block;
        }
      }

      media-preview-chapter-display,
      ::slotted(media-preview-chapter-display) {
        font-size: var(--media-font-size, 13px);
        line-height: 17px;
        min-width: 0;
        visibility: hidden;
        
        transition: min-width 0s, border-radius 0s, margin 0s, padding 0s, visibility 0s;
        transition-delay: calc(var(--media-preview-transition-delay-out, 0s) + var(--media-preview-transition-duration-out, .25s));
        background: var(--media-preview-chapter-background, var(--_preview-background));
        border-radius: var(--media-preview-chapter-border-radius,
          var(--media-preview-border-radius) var(--media-preview-border-radius)
          var(--media-preview-border-radius) var(--media-preview-border-radius));
        padding: var(--media-preview-chapter-padding, 3.5px 9px);
        margin: var(--media-preview-chapter-margin, 0 0 5px);
        text-shadow: var(--media-preview-chapter-text-shadow, 0 0 4px rgb(0 0 0 / .75));
      }

      :host([${b.MEDIA_PREVIEW_IMAGE}]) media-preview-chapter-display,
      :host([${b.MEDIA_PREVIEW_IMAGE}]) ::slotted(media-preview-chapter-display) {
        transition-delay: var(--media-preview-transition-delay-in, .25s);
        border-radius: var(--media-preview-chapter-border-radius, 0);
        padding: var(--media-preview-chapter-padding, 3.5px 9px 0);
        margin: var(--media-preview-chapter-margin, 0);
        min-width: 100%;
      }

      media-preview-chapter-display[${b.MEDIA_PREVIEW_CHAPTER}],
      ::slotted(media-preview-chapter-display[${b.MEDIA_PREVIEW_CHAPTER}]) {
        visibility: visible;
      }

      media-preview-chapter-display:not([aria-valuetext]),
      ::slotted(media-preview-chapter-display:not([aria-valuetext])) {
        display: none;
      }

      media-preview-time-display,
      ::slotted(media-preview-time-display),
      media-time-display,
      ::slotted(media-time-display) {
        font-size: var(--media-font-size, 13px);
        line-height: 17px;
        min-width: 0;
        
        transition: min-width 0s, border-radius 0s;
        transition-delay: calc(var(--media-preview-transition-delay-out, 0s) + var(--media-preview-transition-duration-out, .25s));
        background: var(--media-preview-time-background, var(--_preview-background));
        border-radius: var(--media-preview-time-border-radius,
          var(--media-preview-border-radius) var(--media-preview-border-radius)
          var(--media-preview-border-radius) var(--media-preview-border-radius));
        padding: var(--media-preview-time-padding, 3.5px 9px);
        margin: var(--media-preview-time-margin, 0);
        text-shadow: var(--media-preview-time-text-shadow, 0 0 4px rgb(0 0 0 / .75));
        transform: translateX(min(
          max(calc(50% - var(--_box-width) / 2),
          calc(var(--_box-shift, 0))),
          calc(var(--_box-width) / 2 - 50%)
        ));
      }

      :host([${b.MEDIA_PREVIEW_IMAGE}]) media-preview-time-display,
      :host([${b.MEDIA_PREVIEW_IMAGE}]) ::slotted(media-preview-time-display) {
        transition-delay: var(--media-preview-transition-delay-in, .25s);
        border-radius: var(--media-preview-time-border-radius,
          0 0 var(--media-preview-border-radius) var(--media-preview-border-radius));
        min-width: 100%;
      }

      :host([${b.MEDIA_PREVIEW_TIME}]:hover) {
        --media-time-range-hover-display: block;
      }

      [part~="arrow"],
      ::slotted([part~="arrow"]) {
        display: var(--media-box-arrow-display, inline-block);
        transform: translateX(min(
          max(calc(50% - var(--_box-width) / 2 + var(--media-box-arrow-offset)),
          calc(var(--_box-shift, 0))),
          calc(var(--_box-width) / 2 - 50% - var(--media-box-arrow-offset))
        ));
        
        border-color: transparent;
        border-top-color: var(--media-box-arrow-background, var(--_control-background));
        border-width: var(--media-box-arrow-border-width,
          var(--media-box-arrow-height, 5px) var(--media-box-arrow-width, 6px) 0);
        border-style: solid;
        justify-content: center;
        height: 0;
      }
    </style>
    <div id="preview-rail">
      <slot name="preview" part="box preview-box">
        <media-preview-thumbnail>
          <template shadowrootmode="${wc.shadowRootOptions.mode}">
            ${wc.getTemplateHTML({})}
          </template>
        </media-preview-thumbnail>
        <media-preview-chapter-display></media-preview-chapter-display>
        <media-preview-time-display></media-preview-time-display>
        <slot name="preview-arrow"><div part="arrow"></div></slot>
      </slot>
    </div>
    <div id="current-rail">
      <slot name="current" part="box current-box">
        
      </slot>
    </div>
  `}var mu=(e,t=e.mediaCurrentTime)=>{let n=Number.isFinite(e.mediaSeekableStart)?e.mediaSeekableStart:0,r=Number.isFinite(e.mediaDuration)?e.mediaDuration:e.mediaSeekableEnd;if(Number.isNaN(r))return 0;let i=(t-n)/(r-n);return Math.max(0,Math.min(i,1))},hu=(e,t=e.range.valueAsNumber)=>{let n=Number.isFinite(e.mediaSeekableStart)?e.mediaSeekableStart:0,r=Number.isFinite(e.mediaDuration)?e.mediaDuration:e.mediaSeekableEnd;return Number.isNaN(r)?0:t*(r-n)+n},gu=class extends ao{constructor(){super(),Z(this,Zl),Z(this,eu),Z(this,nu),Z(this,iu),Z(this,ou),Z(this,cu),Z(this,uu),Z(this,Bl,null),Z(this,Vl,void 0),Z(this,Hl,void 0),Z(this,Ul,void 0),Z(this,Wl,void 0),Z(this,Gl,void 0),Z(this,Kl,void 0),Z(this,ql,void 0),Z(this,Jl,void 0),Z(this,Yl,void 0),Z(this,Xl,()=>{zl(this,Zl,Ql).call(this)?X(this,Vl).start():X(this,Vl).stop()}),Z(this,$l,e=>{this.dragging||(He(e)&&(this.range.valueAsNumber=e),X(this,Yl)||this.updateBar())}),this.shadowRoot.querySelector(`#track`).insertAdjacentHTML(`afterbegin`,`<div id="buffered" part="buffered"></div>`),Rl(this,Hl,this.shadowRoot.querySelectorAll(`[part~="box"]`)),Rl(this,Wl,this.shadowRoot.querySelector(`[part~="preview-box"]`)),Rl(this,Gl,this.shadowRoot.querySelector(`[part~="current-box"]`));let e=getComputedStyle(this);Rl(this,Kl,parseInt(e.getPropertyValue(`--media-box-padding-left`))),Rl(this,ql,parseInt(e.getPropertyValue(`--media-box-padding-right`))),Rl(this,Vl,new Il(this.range,X(this,$l),60))}static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_PAUSED,b.MEDIA_DURATION,b.MEDIA_SEEKABLE,b.MEDIA_CURRENT_TIME,b.MEDIA_PREVIEW_IMAGE,b.MEDIA_PREVIEW_TIME,b.MEDIA_PREVIEW_CHAPTER,b.MEDIA_BUFFERED,b.MEDIA_PLAYBACK_RATE,b.MEDIA_LOADING,b.MEDIA_ENDED]}connectedCallback(){var e;super.connectedCallback(),this.range.setAttribute(`aria-label`,E(`seek`)),X(this,Xl).call(this),Rl(this,Bl,this.getRootNode()),(e=X(this,Bl))==null||e.addEventListener(`transitionstart`,this)}disconnectedCallback(){var e;super.disconnectedCallback(),X(this,Vl).stop(),(e=X(this,Bl))==null||e.removeEventListener(`transitionstart`,this),Rl(this,Bl,null)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),t!=n&&(e===b.MEDIA_CURRENT_TIME||e===b.MEDIA_PAUSED||e===b.MEDIA_ENDED||e===b.MEDIA_LOADING||e===b.MEDIA_DURATION||e===b.MEDIA_SEEKABLE?(X(this,Vl).update({start:mu(this),duration:this.mediaSeekableEnd-this.mediaSeekableStart,playbackRate:this.mediaPlaybackRate}),X(this,Xl).call(this),fu(this)):e===b.MEDIA_BUFFERED&&this.updateBufferedBar(),(e===b.MEDIA_DURATION||e===b.MEDIA_SEEKABLE)&&(this.mediaChaptersCues=X(this,Jl),this.updateBar()))}get mediaChaptersCues(){return X(this,Jl)}set mediaChaptersCues(e){Rl(this,Jl,e),this.updateSegments(X(this,Jl)?.map(e=>({start:mu(this,e.startTime),end:mu(this,e.endTime)})))}get mediaPaused(){return N(this,b.MEDIA_PAUSED)}set mediaPaused(e){P(this,b.MEDIA_PAUSED,e)}get mediaLoading(){return N(this,b.MEDIA_LOADING)}set mediaLoading(e){P(this,b.MEDIA_LOADING,e)}get mediaDuration(){return j(this,b.MEDIA_DURATION)}set mediaDuration(e){M(this,b.MEDIA_DURATION,e)}get mediaCurrentTime(){return j(this,b.MEDIA_CURRENT_TIME)}set mediaCurrentTime(e){M(this,b.MEDIA_CURRENT_TIME,e)}get mediaPlaybackRate(){return j(this,b.MEDIA_PLAYBACK_RATE,1)}set mediaPlaybackRate(e){M(this,b.MEDIA_PLAYBACK_RATE,e)}get mediaBuffered(){let e=this.getAttribute(b.MEDIA_BUFFERED);return e?e.split(` `).map(e=>e.split(`:`).map(e=>+e)):[]}set mediaBuffered(e){if(!e){this.removeAttribute(b.MEDIA_BUFFERED);return}let t=e.map(e=>e.join(`:`)).join(` `);this.setAttribute(b.MEDIA_BUFFERED,t)}get mediaSeekable(){let e=this.getAttribute(b.MEDIA_SEEKABLE);if(e)return e.split(`:`).map(e=>+e)}set mediaSeekable(e){e==null?this.removeAttribute(b.MEDIA_SEEKABLE):this.setAttribute(b.MEDIA_SEEKABLE,e.join(`:`))}get mediaSeekableEnd(){let[,e=this.mediaDuration]=this.mediaSeekable??[];return e}get mediaSeekableStart(){let[e=0]=this.mediaSeekable??[];return e}get mediaPreviewImage(){return F(this,b.MEDIA_PREVIEW_IMAGE)}set mediaPreviewImage(e){I(this,b.MEDIA_PREVIEW_IMAGE,e)}get mediaPreviewTime(){return j(this,b.MEDIA_PREVIEW_TIME)}set mediaPreviewTime(e){M(this,b.MEDIA_PREVIEW_TIME,e)}get mediaEnded(){return N(this,b.MEDIA_ENDED)}set mediaEnded(e){P(this,b.MEDIA_ENDED,e)}updateBar(){super.updateBar(),this.updateBufferedBar(),this.updateCurrentBox()}updateBufferedBar(){let e=this.mediaBuffered;if(!e.length)return;let t;if(this.mediaEnded)t=1;else{let n=this.mediaCurrentTime,[,r=this.mediaSeekableStart]=e.find(([e,t])=>e<=n&&n<=t)??[];t=mu(this,r)}let{style:n}=A(this.shadowRoot,`#buffered`);n.setProperty(`width`,`${t*100}%`)}updateCurrentBox(){if(!this.shadowRoot.querySelector(`slot[name="current"]`).assignedElements().length)return;let e=A(this.shadowRoot,`#current-rail`),t=A(this.shadowRoot,`[part~="current-box"]`),n=zl(this,eu,tu).call(this,X(this,Gl)),r=zl(this,nu,ru).call(this,n,this.range.valueAsNumber),i=zl(this,iu,au).call(this,n,this.range.valueAsNumber);e.style.transform=`translateX(${r})`,e.style.setProperty(`--_range-width`,`${n.range.width}`),t.style.setProperty(`--_box-shift`,`${i}`),t.style.setProperty(`--_box-width`,`${n.box.width}px`),t.style.setProperty(`visibility`,`initial`)}handleEvent(e){switch(super.handleEvent(e),e.type){case`input`:zl(this,uu,du).call(this);break;case`pointermove`:zl(this,ou,su).call(this,e);break;case`pointerup`:X(this,Yl)&&Rl(this,Yl,!1);break;case`pointerdown`:Rl(this,Yl,!0);break;case`pointerleave`:zl(this,cu,lu).call(this,null);break;case`transitionstart`:yt(e.target,this)&&setTimeout(()=>X(this,Xl).call(this),0)}}};Bl=new WeakMap,Vl=new WeakMap,Hl=new WeakMap,Ul=new WeakMap,Wl=new WeakMap,Gl=new WeakMap,Kl=new WeakMap,ql=new WeakMap,Jl=new WeakMap,Yl=new WeakMap,Xl=new WeakMap,Zl=new WeakSet,Ql=function(){return this.isConnected&&!this.mediaPaused&&!this.mediaLoading&&!this.mediaEnded&&this.mediaSeekableEnd>0&&Ct(this)},$l=new WeakMap,eu=new WeakSet,tu=function(e){let t=((this.getAttribute(`bounds`)?bt(this,`#${this.getAttribute(`bounds`)}`):this.parentElement)??this).getBoundingClientRect(),n=this.range.getBoundingClientRect(),r=e.offsetWidth;return{box:{width:r,min:-(n.left-t.left-r/2),max:t.right-n.left-r/2},bounds:t,range:n}},nu=new WeakSet,ru=function(e,t){let n=`${t*100}%`,{width:r,min:i,max:a}=e.box;if(!r)return n;if(Number.isNaN(i)||(n=`max(${`calc(1 / var(--_range-width) * 100 * ${i}% + var(--media-box-padding-left))`}, ${n})`),!Number.isNaN(a)){let e=`calc(1 / var(--_range-width) * 100 * ${a}% - var(--media-box-padding-right))`;n=`min(${n}, ${e})`}return n},iu=new WeakSet,au=function(e,t){let{width:n,min:r,max:i}=e.box,a=t*e.range.width;if(a<r+X(this,Kl)){let t=e.range.left-e.bounds.left-X(this,Kl);return`${a-n/2+t}px`}if(a>i-X(this,ql)){let t=e.bounds.right-e.range.right-X(this,ql);return`${a+n/2-t-e.range.width}px`}return 0},ou=new WeakSet,su=function(e){let t=[...X(this,Hl)].some(t=>e.composedPath().includes(t));if(!this.dragging&&(t||!e.composedPath().includes(this))){zl(this,cu,lu).call(this,null);return}let n=this.mediaSeekableEnd;if(!n)return;let r=A(this.shadowRoot,`#preview-rail`),i=A(this.shadowRoot,`[part~="preview-box"]`),a=zl(this,eu,tu).call(this,X(this,Wl)),o=(e.clientX-a.range.left)/a.range.width;o=Math.max(0,Math.min(1,o));let s=zl(this,nu,ru).call(this,a,o),c=zl(this,iu,au).call(this,a,o);r.style.transform=`translateX(${s})`,r.style.setProperty(`--_range-width`,`${a.range.width}`),i.style.setProperty(`--_box-shift`,`${c}`),i.style.setProperty(`--_box-width`,`${a.box.width}px`);let l=Math.round(X(this,Ul))-Math.round(o*n);Math.abs(l)<1&&o>.01&&o<.99||(Rl(this,Ul,o*n),zl(this,cu,lu).call(this,X(this,Ul)))},cu=new WeakSet,lu=function(e){this.dispatchEvent(new O.CustomEvent(v.MEDIA_PREVIEW_REQUEST,{composed:!0,bubbles:!0,detail:e}))},uu=new WeakSet,du=function(){X(this,Vl).stop();let e=hu(this);this.dispatchEvent(new O.CustomEvent(v.MEDIA_SEEK_REQUEST,{composed:!0,bubbles:!0,detail:e}))},gu.shadowRootOptions={mode:`open`},gu.getContainerTemplateHTML=pu,O.customElements.get(`media-time-range`)||O.customElements.define(`media-time-range`,gu);var _u=gu,vu=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},yu=(e,t,n)=>(vu(e,t,`read from private field`),n?n.call(e):t.get(e)),bu=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},xu,Su=1,Cu=e=>e.mediaMuted?0:e.mediaVolume,wu=e=>`${Math.round(e*100)}%`,Tu=class extends ao{constructor(){super(...arguments),bu(this,xu,()=>{let e=this.range.value,t=new O.CustomEvent(v.MEDIA_VOLUME_REQUEST,{composed:!0,bubbles:!0,detail:e});this.dispatchEvent(t)})}static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_VOLUME,b.MEDIA_MUTED,b.MEDIA_VOLUME_UNAVAILABLE]}connectedCallback(){super.connectedCallback(),this.range.setAttribute(`aria-label`,E(`volume`)),this.range.addEventListener(`input`,yu(this,xu))}disconnectedCallback(){this.range.removeEventListener(`input`,yu(this,xu)),super.disconnectedCallback()}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),(e===b.MEDIA_VOLUME||e===b.MEDIA_MUTED)&&(this.range.valueAsNumber=Cu(this),this.range.setAttribute(`aria-valuetext`,wu(this.range.valueAsNumber)),this.updateBar())}get mediaVolume(){return j(this,b.MEDIA_VOLUME,Su)}set mediaVolume(e){M(this,b.MEDIA_VOLUME,e)}get mediaMuted(){return N(this,b.MEDIA_MUTED)}set mediaMuted(e){P(this,b.MEDIA_MUTED,e)}get mediaVolumeUnavailable(){return F(this,b.MEDIA_VOLUME_UNAVAILABLE)}set mediaVolumeUnavailable(e){I(this,b.MEDIA_VOLUME_UNAVAILABLE,e)}};xu=new WeakMap,O.customElements.get(`media-volume-range`)||O.customElements.define(`media-volume-range`,Tu);var Eu=Tu;function Du(e){return`
      <style>
        :host {
          min-width: 4ch;
          padding: var(--media-button-padding, var(--media-control-padding, 10px 5px));
          width: 100%;
          display: grid;
          grid-template-columns: 1fr auto;
          gap: 1rem;
          font-weight: var(--media-button-font-weight, normal);
        }

        #checked-indicator {
          display: none;
        }

        :host([${b.MEDIA_LOOP}]) #checked-indicator {
          display: block;
        }
      </style>
      
      <span id="icon">
     </span>

      <div id="checked-indicator">
        <svg aria-hidden="true" viewBox="0 1 24 24" part="checked-indicator indicator">
          <path d="m10 15.17 9.193-9.191 1.414 1.414-10.606 10.606-6.364-6.364 1.414-1.414 4.95 4.95Z"/>
        </svg>
      </div>
    `}function Ou(){return E(`Loop`)}var ku=class extends K{constructor(){super(...arguments),this.container=null}static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_LOOP]}connectedCallback(){super.connectedCallback(),this.container=this.shadowRoot?.querySelector(`#icon`)||null,this.container&&(this.container.textContent=E(`Loop`))}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===b.MEDIA_LOOP&&this.container&&this.setAttribute(`aria-checked`,this.mediaLoop?`true`:`false`)}get mediaLoop(){return N(this,b.MEDIA_LOOP)}set mediaLoop(e){P(this,b.MEDIA_LOOP,e)}handleClick(){let e=!this.mediaLoop,t=new O.CustomEvent(v.MEDIA_LOOP_REQUEST,{composed:!0,bubbles:!0,detail:e});this.dispatchEvent(t)}};ku.getSlotTemplateHTML=Du,ku.getTooltipContentHTML=Ou,O.customElements.get(`media-loop-button`)||O.customElements.define(`media-loop-button`,ku);var Au=ku;function Q(e){if(typeof e==`boolean`)return e?``:void 0;if(typeof e!=`function`){if(Array.isArray(e)&&e.every(e=>typeof e==`string`||typeof e==`number`||typeof e==`boolean`))return e.join(` `);if(!(typeof e==`object`&&e))return e}}_({tagName:`media-gesture-receiver`,elementClass:Pt,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-container`,elementClass:pn,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-chrome-dialog`,elementClass:Rn,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-keyboard-shortcuts-dialog`,elementClass:Zn,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var ju=_({tagName:`media-controller`,elementClass:Ri,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});_({tagName:`media-tooltip`,elementClass:Hi,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-chrome-button`,elementClass:oa,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-airplay-button`,elementClass:fa,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var Mu=_({tagName:`media-captions-button`,elementClass:xa,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});_({tagName:`media-cast-button`,elementClass:Oa,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-chrome-range`,elementClass:oo,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var Nu=_({tagName:`media-control-bar`,elementClass:ho,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});_({tagName:`media-text-display`,elementClass:wo,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-duration-display`,elementClass:Mo,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-error-dialog`,elementClass:Ko,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var Pu=_({tagName:`media-fullscreen-button`,elementClass:is,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});_({tagName:`media-live-button`,elementClass:ps,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-loading-indicator`,elementClass:Ts,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var Fu=_({tagName:`media-mute-button`,elementClass:Ns,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Iu=_({tagName:`media-pip-button`,elementClass:zs,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Lu=_({tagName:`media-playback-rate-button`,elementClass:Xs,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Ru=_({tagName:`media-play-button`,elementClass:rc,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});_({tagName:`media-poster-image`,elementClass:lc,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-preview-chapter-display`,elementClass:gc,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-preview-thumbnail`,elementClass:wc,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-preview-time-display`,elementClass:jc,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var zu=_({tagName:`media-seek-backward-button`,elementClass:Bc,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Bu=_({tagName:`media-seek-forward-button`,elementClass:Yc,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Vu=_({tagName:`media-time-display`,elementClass:xl,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Hu=_({tagName:`media-time-range`,elementClass:_u,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Uu=_({tagName:`media-volume-range`,elementClass:Eu,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});_({tagName:`media-loop-button`,elementClass:Au,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var $=n(),Wu={noTooltip:!0};function Gu({isMobile:e,showPip:t,showCaptionsButton:n,topChrome:r,extraControls:i,children:a,className:o,style:s}){return(0,$.jsxs)(ju,{className:`ts-player ${o||``}`.trim(),autohide:`2`,style:s,children:[a,(0,$.jsx)(`div`,{slot:`top-chrome`,className:`ts-player-top`,children:r}),(0,$.jsxs)(Nu,{className:`ts-player-bar`,children:[(0,$.jsx)(Ru,{...Wu,className:e?`ts-player-play-lg`:void 0}),(0,$.jsx)(zu,{...Wu,seekOffset:10}),(0,$.jsx)(Bu,{...Wu,seekOffset:10}),(0,$.jsx)(Vu,{showDuration:!0}),(0,$.jsx)(Hu,{}),(0,$.jsx)(Fu,{...Wu}),e?null:(0,$.jsx)(Uu,{}),i,n?(0,$.jsx)(Mu,{...Wu}):null,(0,$.jsx)(Lu,{...Wu,rates:[.75,1,1.25,1.5,2]}),t?(0,$.jsx)(Iu,{...Wu}):null,(0,$.jsx)(Pu,{...Wu})]})]})}var Ku=e=>!!(e.canPlayType(`application/vnd.apple.mpegurl`)||e.canPlayType(`application/x-mpegURL`));function qu({enabled:e,open:t,videoRef:n,videoEpoch:r,src:i,playbackRate:a=1,onLoading:o,onError:s,onNotSupported:c,onSubtitleTracks:l,onSubtitleTrack:u,onDuration:d}){let m=(0,g.useRef)(null),h=(0,g.useRef)(o),ee=(0,g.useRef)(s),te=(0,g.useRef)(c),ne=(0,g.useRef)(l),re=(0,g.useRef)(u),ie=(0,g.useRef)(d);return(0,g.useEffect)(()=>{h.current=o,ee.current=s,te.current=c,ne.current=l,re.current=u,ie.current=d}),(0,g.useEffect)(()=>{let r=n.current;if(!e||!t||!r||!i)return;let o=null,s=!1;h.current?.(!0),ne.current?.([]);let c=()=>{let e=(o?.subtitleTracks||[]).map((e,t)=>({id:t,name:e.name,lang:e.lang}));ne.current?.(e)};if(Te.isSupported()){let e=f();o=new Te({xhrSetup:t=>{e&&t.setRequestHeader(`Authorization`,e)}}),m.current=o,o.on(Te.Events.MANIFEST_PARSED,()=>{c(),re.current?.(o?.subtitleTrack??-1),r.playbackRate=a,r.play().catch(()=>{})}),o.on(Te.Events.LEVEL_LOADED,(e,t)=>{let n=t.details?.totalduration;typeof n==`number`&&ie.current?.(n)}),o.on(Te.Events.SUBTITLE_TRACKS_UPDATED,c),o.on(Te.Events.SUBTITLE_TRACK_SWITCH,(e,t)=>re.current?.(t.id)),o.on(Te.Events.ERROR,(e,t)=>{t.fatal&&o&&(t.type===Te.ErrorTypes.NETWORK_ERROR?o.startLoad():t.type===Te.ErrorTypes.MEDIA_ERROR?o.recoverMediaError():(o.stopLoad(),h.current?.(!1),ee.current?.()))}),o.loadSource(i),o.attachMedia(r)}else Ku(r)?(s=!0,r.src=p(i),r.load(),r.playbackRate=a,r.play().catch(()=>{})):te.current?.();return()=>{m.current===o&&(m.current=null),o?.destroy(),s&&(r.pause(),r.removeAttribute(`src`),r.load()),ne.current?.([]),re.current?.(-1)}},[e,t,n,r,i]),m}var Ju=5e3,Yu=5;function Xu({enabled:e,hash:t,fileIndex:n,title:r,initialTimecode:i=0,videoRef:a,onViewedChange:o}){let s=(0,g.useRef)(0),c=(0,g.useRef)(!1),l=(0,g.useRef)(o);(0,g.useEffect)(()=>{l.current=o},[o]),(0,g.useEffect)(()=>{c.current=!1},[t,n,i]);let u=(0,g.useCallback)(async i=>{if(e&&t&&n!=null)try{await ge(t,n,i),me({hash:t,fileIndex:n,title:r||t,fileName:r||String(n),timecode:i}),l.current?.()}catch{}},[e,t,n,r]);return{flushTimecode:(0,g.useCallback)(()=>{let t=a.current;t&&e&&u(t.currentTime)},[a,e,u]),onTimeUpdate:(0,g.useCallback)(()=>{let t=a.current;if(!t||!e)return;let n=Date.now();n-s.current<Ju||(s.current=n,u(t.currentTime))},[a,e,u]),applyResumeIfNeeded:(0,g.useCallback)(()=>{let e=a.current;if(!e||c.current)return;if(!(i>Yu)){c.current=!0;return}let t=e.duration;Number.isFinite(t)&&t>0&&i>=t-Yu||(e.currentTime=i),c.current=!0},[a,i]),saveTimecode:u}}var Zu=3e4;function Qu({tip:e,children:t}){return(0,$.jsx)(`span`,{className:`inline-flex min-w-0 max-w-full items-center gap-1`,title:e,children:t})}function $u(e){switch(e.split(`?`)[0].split(`.`).pop()?.toLowerCase()){case`mp4`:return`video/mp4`;case`ogg`:case`ogv`:return`video/ogg`;case`webm`:return`video/webm`;default:return``}}var ed=e=>!!(e.canPlayType(`application/vnd.apple.mpegurl`)||e.canPlayType(`application/x-mpegURL`)),td=()=>typeof document<`u`&&`pictureInPictureEnabled`in document&&!!document.pictureInPictureEnabled;function nd({videoSrc:e,downloadSrc:n=e,title:r,onNotSupported:f,hls:h=!1,heartbeatSrc:me=``,showTrigger:ge=!0,inlineTrigger:Ee=!1,inlineTriggerPrimary:Oe=!1,initiallyOpen:ke=!1,onClose:Ae,captionSrc:je,hash:_,fileIndex:Me,initialTimecode:v=0,trackTimecode:y=!1,onViewedChange:Ne,audioTracks:Pe=[],audioIndex:b=0,onAudioIndexChange:Fe}){let{t:x}=ne(),S=ye(),{setImmersive:Ie}=u(),Le=(0,g.useRef)(null),C=(0,g.useRef)(f),w=(0,g.useRef)(null),[T,Re]=(0,g.useState)(ke),[ze,Be]=(0,g.useState)(0),[Ve,He]=(0,g.useState)(!0),[Ue,We]=(0,g.useState)(!1),[Ge,Ke]=(0,g.useState)(!1),[qe,Je]=(0,g.useState)(!1),[Ye,Xe]=(0,g.useState)([]),[E,Ze]=(0,g.useState)(!1),[Qe,$e]=(0,g.useState)(b),[D,et]=(0,g.useState)(e),[tt,nt]=(0,g.useState)(Pe),{data:rt}=we(T?_:void 0),it=Ce(_,{enabled:T&&!!_,fast:!1}),at=!!(_&&Me!=null&&y),ot=h&&!!_&&Me!=null&&tt.length>1,st=td(),O=h?Ye.length>0:!!je,k=S?pe:qe?he:ue,ct=S?void 0:qe?`min(86dvh, calc(100dvh - 4rem))`:`min(72dvh, 40rem)`,lt=(0,g.useCallback)(e=>{Le.current=e,Be(e=>e+1)},[]);(0,g.useEffect)(()=>{C.current=f},[f]),(0,g.useEffect)(()=>{et(e),$e(b),nt(Pe),We(!1),He(!0)},[e,b,Pe]),qu({enabled:h,open:T,videoRef:Le,videoEpoch:ze,src:D,onLoading:He,onError:()=>We(!0),onNotSupported:()=>C.current?.(),onSubtitleTracks:Xe});let{flushTimecode:ut,onTimeUpdate:dt,applyResumeIfNeeded:ft}=Xu({enabled:at,hash:_,fileIndex:Me,title:r,initialTimecode:v,videoRef:Le,onViewedChange:Ne}),pt=(0,g.useCallback)(()=>Re(!0),[]),mt=(0,g.useCallback)(()=>{ut(),Re(!1),We(!1),Ze(!1),Je(!1),Ae?.()},[ut,Ae]),ht=c({isOpen:T,onOpenChange:e=>{e?Re(!0):mt()}});d(T),(0,g.useEffect)(()=>{if(!S||!T)Ie(!1);else return Ie(!0),()=>Ie(!1)},[S,T,Ie]),(0,g.useEffect)(()=>{let e=document.createElement(`video`);(h?Te.isSupported()||ed(e):e.canPlayType($u(D)))||C.current?.()},[h,D]),(0,g.useEffect)(()=>{let e=Le.current;if(T&&!h&&e)return He(!0),e.src=p(D),e.load(),e.play().catch(()=>{}),()=>{e.pause(),e.removeAttribute(`src`),e.load()}},[T,h,ze,D]),(0,g.useEffect)(()=>{if(!T||!me)return;let e=window.setInterval(()=>{m(me,{cache:`no-store`}).catch(()=>{})},Zu);return()=>window.clearInterval(e)},[me,T]),(0,g.useEffect)(()=>{let e=()=>Ke(!!document.fullscreenElement);return document.addEventListener(`fullscreenchange`,e),()=>document.removeEventListener(`fullscreenchange`,e)},[]),(0,g.useEffect)(()=>{let e=Le.current;if(!e||w.current==null)return;let t=w.current,n=()=>{w.current!=null&&(e.currentTime=t,w.current=null,e.play().catch(()=>{}))};return e.addEventListener(`loadedmetadata`,n,{once:!0}),()=>e.removeEventListener(`loadedmetadata`,n)},[ze,D]),(0,g.useEffect)(()=>{if(!T||!h||!_||Me==null)return;if(Pe.length){nt(Pe);return}let e=!1;return l.get(fe(_,Me)).then(({data:t})=>{e||nt(_e(t))}).catch(()=>{}),()=>{e=!0}},[T,h,_,Me,Pe]);let gt=e=>{if(!_||Me==null)return;let t=Le.current;t&&(w.current=t.currentTime),$e(e),et(Se(_,Me,e)),Ze(!1),Fe?.(e)},_t=oe,vt=rt?.download_speed??0,yt=rt?.active_peers,bt=rt?.total_peers??0,xt=rt?.connected_seeders??0,St=yt==null?`—`:`${yt}/${bt}`,Ct=yt==null?`—`:`↑${xt}`,wt=ce(it.Filled,it.Capacity)??`—`,A=_==null?null:(0,$.jsxs)(`div`,{className:`mt-0.5 flex min-w-0 flex-wrap items-center gap-x-2.5 gap-y-1 text-[11px] leading-none text-white/80 drop-shadow sm:text-xs`,"aria-label":`${x(`DownloadSpeed`)}, ${x(`Peers`)}, ${x(`CacheFilled`)}`,children:[(0,$.jsxs)(Qu,{tip:x(`DownloadSpeed`),children:[(0,$.jsx)(ve,{className:`size-3 shrink-0`,strokeWidth:2.25,"aria-hidden":!0}),(0,$.jsx)(`span`,{className:`tabular-nums`,children:re(vt)})]}),(0,$.jsxs)(Qu,{tip:x(`Peers`),children:[(0,$.jsx)(se,{className:`size-3 shrink-0`,strokeWidth:2,"aria-hidden":!0}),(0,$.jsx)(`span`,{className:`tabular-nums`,children:St}),(0,$.jsx)(`span`,{className:`tabular-nums text-white/70`,children:Ct})]}),(0,$.jsxs)(Qu,{tip:x(`CacheHint`),children:[(0,$.jsx)(ee,{className:`size-3 shrink-0`,strokeWidth:2,"aria-hidden":!0}),(0,$.jsx)(`span`,{className:`min-w-0 truncate tabular-nums`,children:wt})]})]}),Tt=ot?(0,$.jsxs)(o,{isOpen:E,onOpenChange:Ze,children:[(0,$.jsx)(o.Trigger,{children:(0,$.jsx)(t,{isIconOnly:!0,variant:`ghost`,className:`${be} text-white hover-fine:bg-white/15`,"aria-label":x(`SelectAudioTrack`),children:(0,$.jsx)(xe,{..._t,"aria-hidden":!0})})}),(0,$.jsxs)(o.Content,{placement:`bottom end`,className:`max-h-72 w-72 overflow-y-auto border border-white/10 bg-neutral-950/95 p-1 backdrop-blur-xl`,children:[(0,$.jsx)(`p`,{className:`px-2 py-1.5 text-[11px] font-semibold tracking-wide text-muted uppercase`,children:x(`SelectAudioTrack`)}),tt.map((e,n)=>{let{title:r,meta:i}=de(e,n);return(0,$.jsx)(t,{variant:n===Qe?`secondary`:`ghost`,className:`h-auto w-full justify-start gap-2 py-2`,onPress:()=>gt(n),children:(0,$.jsxs)(`span`,{className:`min-w-0 flex-1 text-left`,children:[(0,$.jsx)(`span`,{className:`block truncate text-sm font-medium`,children:r}),i?(0,$.jsx)(`span`,{className:`mt-0.5 block truncate text-xs text-muted`,children:i}):null]})},n)})]})]}):null,Et=S?null:(0,$.jsx)(t,{isIconOnly:!0,variant:`ghost`,className:`${be} text-white hover-fine:bg-white/15`,"aria-label":x(qe?`ExitFullscreen`:`ExpandPlayer`),onPress:()=>Je(e=>!e),children:qe?(0,$.jsx)(De,{..._t,"aria-hidden":!0}):(0,$.jsx)(te,{..._t,"aria-hidden":!0})});return(0,$.jsxs)($.Fragment,{children:[ge&&(Ee?(0,$.jsxs)(t,{variant:Oe?`primary`:`secondary`,size:`sm`,onPress:pt,className:Oe?`min-h-10 shrink-0 px-3`:`min-h-10 min-w-[72px] max-w-full flex-1`,children:[Oe?(0,$.jsx)(le,{...ae,fill:`currentColor`,"aria-hidden":!0}):null,x(`Play`)]}):(0,$.jsxs)(t,{variant:`secondary`,onPress:pt,children:[(0,$.jsx)(le,{...ae,"aria-hidden":!0}),x(`Play`)]})),(0,$.jsx)(i,{state:ht,children:(0,$.jsx)(i.Backdrop,{isDismissable:!Ge,className:`bg-black/75 backdrop-blur-sm`,children:(0,$.jsx)(i.Container,{size:S?`full`:`lg`,scroll:`inside`,placement:S?`center`:`auto`,className:S?`ts-player-modal-full p-0`:`p-4 sm:p-5`,children:(0,$.jsx)(i.Dialog,{className:S?`overflow-hidden rounded-none border-0 bg-black text-white shadow-none`:`overflow-hidden border border-white/10 bg-[#0a0e0c] text-white shadow-2xl shadow-black/50`,style:k,children:(0,$.jsxs)(i.Body,{className:S?`flex h-full min-h-0 flex-col gap-0 p-0`:`gap-0 p-0`,children:[Ue?(0,$.jsxs)(`div`,{className:`space-y-3 p-4`,children:[(0,$.jsx)(s,{status:`danger`,children:x(`PlaybackError`)}),(0,$.jsx)(t,{variant:`secondary`,onPress:()=>window.open(n,`_blank`,`noopener,noreferrer`),children:x(`OpenLink`)})]}):(0,$.jsx)(Gu,{isMobile:S,showPip:st,showCaptionsButton:O,className:S?`ts-player--mobile`:`ts-player--desktop`,style:S?{width:`100%`,height:`100%`}:{width:`100%`,maxHeight:ct,aspectRatio:`16 / 9`},topChrome:(0,$.jsxs)(`div`,{className:`flex items-start gap-2 px-3 pb-10 pt-[max(0.75rem,env(safe-area-inset-top))] pl-[max(0.75rem,env(safe-area-inset-left))] pr-[max(0.75rem,env(safe-area-inset-right))]`,children:[(0,$.jsxs)(`div`,{className:`min-w-0 flex-1`,children:[(0,$.jsx)(`p`,{className:`truncate text-sm font-semibold text-white drop-shadow`,title:r||x(`Play`),children:r||x(`Play`)}),A]}),Tt,Et,(0,$.jsx)(t,{isIconOnly:!0,variant:`ghost`,className:`${be} text-white hover-fine:bg-white/15`,"aria-label":x(`Close`),onPress:mt,children:(0,$.jsx)(ie,{..._t,"aria-hidden":!0})})]}),extraControls:null,children:(0,$.jsx)(`video`,{slot:`media`,ref:lt,autoPlay:!0,playsInline:!0,crossOrigin:`anonymous`,className:`ts-player-video`,onTimeUpdate:()=>{dt(),Ve&&He(!1)},onLoadedMetadata:()=>{He(!1),We(!1),ft()},onWaiting:()=>He(!0),onCanPlay:()=>{He(!1),We(!1)},onPlaying:()=>He(!1),onPause:ut,onError:()=>{He(!1),We(!0)},children:!h&&je?(0,$.jsx)(`track`,{kind:`captions`,src:p(je),srcLang:`und`,label:`Captions`,default:!0}):null},D)}),Ve&&!Ue?(0,$.jsx)(`div`,{className:`pointer-events-none absolute inset-0 z-10 grid place-items-center bg-black/55`,children:(0,$.jsxs)(`div`,{className:`flex flex-col items-center gap-3 px-4 text-center`,children:[(0,$.jsx)(a,{size:`lg`,color:`current`,className:`text-accent`}),(0,$.jsx)(`p`,{className:`text-sm font-medium text-white/90`,children:x(`Buffering`)})]})}):null]})})})})})]})}export{nd as default};