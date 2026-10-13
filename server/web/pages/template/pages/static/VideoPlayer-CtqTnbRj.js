import{i as e}from"./rolldown-runtime-8BhlS34s.js";import{E as t,N as n,_ as r,f as i,g as a,j as o,k as s,r as c}from"./heroui-C7OCMOHJ.js";import{i as l}from"./vendor-Dy7IQDr1.js";import{n as u,r as d}from"./ModalOpenContext-CXMCD-lH.js";import{a as f,l as p,n as m}from"./authCredentials-BZhgIdwZ.js";import{t as h}from"./createLucideIcon-BrMNLKI3.js";import{n as ee,t as te}from"./maximize-2-jc_lEpTl.js";import{t as ne}from"./useTranslation-D7c2ATyD.js";import{At as re,Dt as ie,H as ae,K as oe,Ot as se,Vt as ce,Wt as le,_ as ue,a as de,b as fe,d as pe,et as me,g as he,i as ge,jt as _e,nn as ve,ot as ye,s as be,tt as xe,y as Se}from"./index-drS62LXu.js";import{r as Ce,t as we}from"./useTorrentDetail-BIo_8YC9.js";import{t as Te}from"./hls-DbOvCaf2.js";var Ee={name:`minimize-2`,size:24,node:[[`path`,{d:`m14 10 7-7`,key:`oa77jy`}],[`path`,{d:`M20 10h-6V4`,key:`mjg0md`}],[`path`,{d:`m3 21 7-7`,key:`tjx5ai`}],[`path`,{d:`M4 14h6v6`,key:`rmj7iw`}]]};Ee.node;var De=h(Ee),g=e(n(),1),Oe=new Set([`style`,`children`,`ref`,`key`,`suppressContentEditableWarning`,`suppressHydrationWarning`,`dangerouslySetInnerHTML`]),ke={className:`class`,htmlFor:`for`};function Ae(e){return e.toLowerCase()}function je(e){if(typeof e==`boolean`)return e?``:void 0;if(typeof e!=`function`&&!(typeof e==`object`&&e))return e}function _({react:e,tagName:t,elementClass:n,events:r,displayName:i,defaultProps:a,toAttributeName:o=Ae,toAttributeValue:s=je}){let c=Number.parseInt(e.version)>=19,l=e.forwardRef((i,l)=>{let u=e.useRef(null),d=e.useRef(new Map),f={},p={},m={},h={};for(let[e,t]of Object.entries(i)){if(Oe.has(e)){m[e]=t;continue}let r=o(ke[e]??e);if(n.prototype&&e in n.prototype&&!(e in(globalThis.HTMLElement?.prototype??{}))&&!n.observedAttributes?.some(e=>e===r)){h[e]=t;continue}if(e.startsWith(`on`)){f[e]=t;continue}let i=s(t);r&&i!=null&&(p[r]=String(i),c||(m[r]=i)),r&&c&&(m[r]=i===je(t)?t:i)}if(typeof window<`u`){for(let t in f){let n=f[t],i=t.endsWith(`Capture`),a=(r?.[t]??t.slice(2).toLowerCase()).slice(0,i?-7:void 0);e.useLayoutEffect(()=>{let e=u?.current;if(e&&typeof n==`function`)return e.addEventListener(a,n,i),()=>{e.removeEventListener(a,n,i)}},[u?.current,n])}e.useLayoutEffect(()=>{if(u.current===null)return;let e=new Map;for(let t in h)Me(u.current,t,h[t]),d.current.delete(t),e.set(t,h[t]);for(let[e,t]of d.current)Me(u.current,e,void 0);d.current=e})}if(typeof window>`u`&&n?.getTemplateHTML&&n?.shadowRootOptions){let{mode:t,delegatesFocus:r}=n.shadowRootOptions;m.children=[e.createElement(`template`,{shadowrootmode:t,shadowrootdelegatesfocus:r,dangerouslySetInnerHTML:{__html:n.getTemplateHTML(p,i)},key:`ce-la-react-ssr-template-shadow-root`}),m.children]}return e.createElement(t,{...a,...m,ref:e.useCallback(e=>{u.current=e,typeof l==`function`?l(e):l!==null&&(l.current=e)},[l])},m.children)});return l.displayName=i??n.name,l}function Me(e,t,n){e[t]=n,n==null&&t in(globalThis.HTMLElement?.prototype??{})&&e.removeAttribute(t)}var v={MEDIA_PLAY_REQUEST:`mediaplayrequest`,MEDIA_PAUSE_REQUEST:`mediapauserequest`,MEDIA_MUTE_REQUEST:`mediamuterequest`,MEDIA_UNMUTE_REQUEST:`mediaunmuterequest`,MEDIA_LOOP_REQUEST:`medialooprequest`,MEDIA_VOLUME_REQUEST:`mediavolumerequest`,MEDIA_SEEK_REQUEST:`mediaseekrequest`,MEDIA_AIRPLAY_REQUEST:`mediaairplayrequest`,MEDIA_ENTER_FULLSCREEN_REQUEST:`mediaenterfullscreenrequest`,MEDIA_EXIT_FULLSCREEN_REQUEST:`mediaexitfullscreenrequest`,MEDIA_PREVIEW_REQUEST:`mediapreviewrequest`,MEDIA_ENTER_PIP_REQUEST:`mediaenterpiprequest`,MEDIA_EXIT_PIP_REQUEST:`mediaexitpiprequest`,MEDIA_ENTER_CAST_REQUEST:`mediaentercastrequest`,MEDIA_EXIT_CAST_REQUEST:`mediaexitcastrequest`,MEDIA_SHOW_TEXT_TRACKS_REQUEST:`mediashowtexttracksrequest`,MEDIA_HIDE_TEXT_TRACKS_REQUEST:`mediahidetexttracksrequest`,MEDIA_SHOW_SUBTITLES_REQUEST:`mediashowsubtitlesrequest`,MEDIA_DISABLE_SUBTITLES_REQUEST:`mediadisablesubtitlesrequest`,MEDIA_TOGGLE_SUBTITLES_REQUEST:`mediatogglesubtitlesrequest`,MEDIA_PLAYBACK_RATE_REQUEST:`mediaplaybackraterequest`,MEDIA_RENDITION_REQUEST:`mediarenditionrequest`,MEDIA_AUDIO_TRACK_REQUEST:`mediaaudiotrackrequest`,MEDIA_SEEK_TO_LIVE_REQUEST:`mediaseektoliverequest`,REGISTER_MEDIA_STATE_RECEIVER:`registermediastatereceiver`,UNREGISTER_MEDIA_STATE_RECEIVER:`unregistermediastatereceiver`},y={MEDIA_CHROME_ATTRIBUTES:`mediachromeattributes`,MEDIA_CONTROLLER:`mediacontroller`},Ne={MEDIA_AIRPLAY_UNAVAILABLE:`mediaAirplayUnavailable`,MEDIA_AUDIO_TRACK_ENABLED:`mediaAudioTrackEnabled`,MEDIA_AUDIO_TRACK_LIST:`mediaAudioTrackList`,MEDIA_AUDIO_TRACK_UNAVAILABLE:`mediaAudioTrackUnavailable`,MEDIA_BUFFERED:`mediaBuffered`,MEDIA_CAST_UNAVAILABLE:`mediaCastUnavailable`,MEDIA_CHAPTERS_CUES:`mediaChaptersCues`,MEDIA_CURRENT_TIME:`mediaCurrentTime`,MEDIA_DURATION:`mediaDuration`,MEDIA_ENDED:`mediaEnded`,MEDIA_ERROR:`mediaError`,MEDIA_ERROR_CODE:`mediaErrorCode`,MEDIA_ERROR_MESSAGE:`mediaErrorMessage`,MEDIA_FULLSCREEN_UNAVAILABLE:`mediaFullscreenUnavailable`,MEDIA_HAS_PLAYED:`mediaHasPlayed`,MEDIA_HEIGHT:`mediaHeight`,MEDIA_IS_AIRPLAYING:`mediaIsAirplaying`,MEDIA_IS_CASTING:`mediaIsCasting`,MEDIA_IS_FULLSCREEN:`mediaIsFullscreen`,MEDIA_IS_PIP:`mediaIsPip`,MEDIA_LOADING:`mediaLoading`,MEDIA_MUTED:`mediaMuted`,MEDIA_LOOP:`mediaLoop`,MEDIA_PAUSED:`mediaPaused`,MEDIA_PIP_UNAVAILABLE:`mediaPipUnavailable`,MEDIA_PLAYBACK_RATE:`mediaPlaybackRate`,MEDIA_PREVIEW_CHAPTER:`mediaPreviewChapter`,MEDIA_PREVIEW_COORDS:`mediaPreviewCoords`,MEDIA_PREVIEW_IMAGE:`mediaPreviewImage`,MEDIA_PREVIEW_TIME:`mediaPreviewTime`,MEDIA_RENDITION_LIST:`mediaRenditionList`,MEDIA_RENDITION_SELECTED:`mediaRenditionSelected`,MEDIA_RENDITION_UNAVAILABLE:`mediaRenditionUnavailable`,MEDIA_SEEKABLE:`mediaSeekable`,MEDIA_STREAM_TYPE:`mediaStreamType`,MEDIA_SUBTITLES_LIST:`mediaSubtitlesList`,MEDIA_SUBTITLES_SHOWING:`mediaSubtitlesShowing`,MEDIA_TARGET_LIVE_WINDOW:`mediaTargetLiveWindow`,MEDIA_TIME_IS_LIVE:`mediaTimeIsLive`,MEDIA_VOLUME:`mediaVolume`,MEDIA_VOLUME_LEVEL:`mediaVolumeLevel`,MEDIA_VOLUME_UNAVAILABLE:`mediaVolumeUnavailable`,MEDIA_LANG:`mediaLang`,MEDIA_WIDTH:`mediaWidth`},Pe=Object.entries(Ne),b=Pe.reduce((e,[t,n])=>(e[t]=n.toLowerCase(),e),{}),Fe=Pe.reduce((e,[t,n])=>(e[t]=n.toLowerCase(),e),{USER_INACTIVE_CHANGE:`userinactivechange`,BREAKPOINTS_CHANGE:`breakpointchange`,BREAKPOINTS_COMPUTED:`breakpointscomputed`});Object.entries(Fe).reduce((e,[t,n])=>{let r=b[t];return r&&(e[n]=r),e},{userinactivechange:`userinactive`});var x=Object.entries(b).reduce((e,[t,n])=>{let r=Fe[t];return r&&(e[n]=r),e},{userinactive:`userinactivechange`}),S={SUBTITLES:`subtitles`,CAPTIONS:`captions`,DESCRIPTIONS:`descriptions`,CHAPTERS:`chapters`,METADATA:`metadata`},Ie={DISABLED:`disabled`,HIDDEN:`hidden`,SHOWING:`showing`},Le={MOUSE:`mouse`,PEN:`pen`,TOUCH:`touch`},C={UNAVAILABLE:`unavailable`,UNSUPPORTED:`unsupported`},w={LIVE:`live`,ON_DEMAND:`on-demand`,UNKNOWN:`unknown`},T={INLINE:`inline`,FULLSCREEN:`fullscreen`,PICTURE_IN_PICTURE:`picture-in-picture`};function Re(e){return e?.map(ze).join(` `)}function ze(e){if(e){let{id:t,width:n,height:r}=e;return[t,n,r].filter(e=>e!=null).join(`:`)}}function Be(e){return e?.map(Ve).join(` `)}function Ve(e){if(e){let{id:t,kind:n,language:r,label:i}=e;return[t,n,r,i].filter(e=>e!=null).join(`:`)}}function He(e){return typeof e==`number`&&!Number.isNaN(e)&&Number.isFinite(e)}var Ue=e=>new Promise(t=>setTimeout(t,e)),We={en:{"Start airplay":`Start airplay`,"Stop airplay":`Stop airplay`,Audio:`Audio`,Captions:`Captions`,"Enable captions":`Enable captions`,"Disable captions":`Disable captions`,"Start casting":`Start casting`,"Stop casting":`Stop casting`,"Enter fullscreen mode":`Enter fullscreen mode`,"Exit fullscreen mode":`Exit fullscreen mode`,Mute:`Mute`,Unmute:`Unmute`,Loop:`Loop`,"Enter picture in picture mode":`Enter picture in picture mode`,"Exit picture in picture mode":`Exit picture in picture mode`,Play:`Play`,Pause:`Pause`,"Playback rate":`Playback rate`,"Playback rate {playbackRate}":`Playback rate {playbackRate}`,Quality:`Quality`,"Seek backward":`Seek backward`,"Seek forward":`Seek forward`,Settings:`Settings`,Auto:`Auto`,"audio player":`audio player`,"video player":`video player`,volume:`volume`,seek:`seek`,"closed captions":`closed captions`,"current playback rate":`current playback rate`,"playback time":`playback time`,"media loading":`media loading`,settings:`settings`,"audio tracks":`audio tracks`,quality:`quality`,play:`play`,pause:`pause`,mute:`mute`,unmute:`unmute`,"chapter: {chapterName}":`chapter: {chapterName}`,live:`live`,Off:`Off`,"start airplay":`start airplay`,"stop airplay":`stop airplay`,"start casting":`start casting`,"stop casting":`stop casting`,"enter fullscreen mode":`enter fullscreen mode`,"exit fullscreen mode":`exit fullscreen mode`,"enter picture in picture mode":`enter picture in picture mode`,"exit picture in picture mode":`exit picture in picture mode`,"seek to live":`seek to live`,"playing live":`playing live`,"seek back {seekOffset} seconds":`seek back {seekOffset} seconds`,"seek forward {seekOffset} seconds":`seek forward {seekOffset} seconds`,"Network Error":`Network Error`,"Decode Error":`Decode Error`,"Source Not Supported":`Source Not Supported`,"Encryption Error":`Encryption Error`,"A network error caused the media download to fail.":`A network error caused the media download to fail.`,"A media error caused playback to be aborted. The media could be corrupt or your browser does not support this format.":`A media error caused playback to be aborted. The media could be corrupt or your browser does not support this format.`,"An unsupported error occurred. The server or network failed, or your browser does not support this format.":`An unsupported error occurred. The server or network failed, or your browser does not support this format.`,"The media is encrypted and there are no keys to decrypt it.":`The media is encrypted and there are no keys to decrypt it.`,hour:`hour`,hours:`hours`,minute:`minute`,minutes:`minutes`,second:`second`,seconds:`seconds`,"{time} remaining":`{time} remaining`,"{currentTime} of {totalTime}":`{currentTime} of {totalTime}`,"video not loaded, unknown time.":`video not loaded, unknown time.`}},Ge=globalThis.navigator?.language||`en`,Ke=e=>{Ge=e},qe=e=>{let[t]=Ge.split(`-`);return We[Ge]?.[e]||We[t]?.[e]||We.en?.[e]||e},Je=()=>{let[e]=Ge.split(`-`);return We[Ge]?Ge:We[e]?e:`en`},E=(e,t={})=>qe(e).replace(/\{(\w+)\}/g,(e,n)=>n in t?String(t[n]):`{${n}}`),Ye=[{singular:`hour`,plural:`hours`},{singular:`minute`,plural:`minutes`},{singular:`second`,plural:`seconds`}],Xe=(e,t)=>`${e} ${E(e===1?Ye[t].singular:Ye[t].plural)}`,Ze=e=>{if(!He(e))return``;let t=Math.abs(e),n=t!==e,r=new Date(0,0,0,0,0,t,0),i=[r.getHours(),r.getMinutes(),r.getSeconds()].map((e,t)=>e&&Xe(e,t)).filter(e=>e).join(`, `);return n?E(`{time} remaining`,{time:i}):t===0?Xe(0,2):i};function Qe(e,t){let n=!1;e<0&&(n=!0,e=0-e),e=e<0?0:e;let r=Math.floor(e%60),i=Math.floor(e/60%60),a=Math.floor(e/3600),o=Math.floor(t/60%60),s=Math.floor(t/3600);return(isNaN(e)||e===1/0)&&(a=i=r=`0`),a=a>0||s>0?a+`:`:``,i=((a||o>=10)&&i<10?`0`+i:i)+`:`,r=r<10?`0`+r:r,(n?`-`:``)+a+i+r}Object.freeze({length:0,start(e){let t=e>>>0;if(t>=this.length)throw new DOMException(`Failed to execute 'start' on 'TimeRanges': The index provided (${t}) is greater than or equal to the maximum bound (${this.length}).`);return 0},end(e){let t=e>>>0;if(t>=this.length)throw new DOMException(`Failed to execute 'end' on 'TimeRanges': The index provided (${t}) is greater than or equal to the maximum bound (${this.length}).`);return 0}});var $e=class{addEventListener(){}removeEventListener(){}dispatchEvent(){return!0}},et=class extends $e{},tt=class extends et{constructor(){super(...arguments),this.role=null}},nt=class{observe(){}unobserve(){}disconnect(){}},rt={createElement:function(){return new it.HTMLElement},createElementNS:function(){return new it.HTMLElement},addEventListener(){},removeEventListener(){},dispatchEvent(e){return!1}},it={ResizeObserver:nt,document:rt,Node:et,Element:tt,HTMLElement:class extends tt{constructor(){super(...arguments),this.innerHTML=``}get content(){return new it.DocumentFragment}},DocumentFragment:class extends $e{},customElements:{get:function(){},define:function(){},whenDefined:function(){}},localStorage:{getItem(e){return null},setItem(e,t){},removeItem(e){}},CustomEvent:function(){},getComputedStyle:function(){},navigator:{languages:[],get userAgent(){return``}},matchMedia(e){return{matches:!1,media:e}},DOMParser:class{parseFromString(e,t){return{body:{textContent:e}}}}},at=`global`in globalThis&&(globalThis==null?void 0:globalThis.global)===globalThis||typeof window>`u`||window.customElements===void 0,ot=Object.keys(it).every(e=>e in globalThis),D=at&&!ot?it:globalThis,O=at&&!ot?rt:globalThis.document,st=new WeakMap,ct=e=>{let t=st.get(e);return t||st.set(e,t=new Set),t},lt=new D.ResizeObserver(e=>{for(let t of e)for(let e of ct(t.target))e(t)});function ut(e,t){ct(e).add(t),lt.observe(e)}function dt(e,t){let n=ct(e);n.delete(t),n.size||lt.unobserve(e)}function ft(e){let t={};for(let n of e)t[n.name]=n.value;return t}function pt(e){return mt(e)??yt(e,`media-controller`)}function mt(e){let{MEDIA_CONTROLLER:t}=y,n=e.getAttribute(t);if(n)return xt(e)?.getElementById(n)}var ht=(e,t,n=`.value`)=>{let r=e.querySelector(n);r&&(r.textContent=t)},gt=(e,t)=>{let n=`slot[name="${t}"]`,r=e.shadowRoot.querySelector(n);return r?r.children:[]},_t=(e,t)=>gt(e,t)[0],vt=(e,t)=>!e||!t?!1:e?.contains(t)?!0:vt(e,t.getRootNode().host),yt=(e,t)=>e?e.closest(t)||yt(e.getRootNode().host,t):null;function bt(e=document){let t=e?.activeElement;return t?bt(t.shadowRoot)??t:null}function xt(e){let t=(e?.getRootNode)?.call(e);return t instanceof ShadowRoot||t instanceof Document?t:null}function St(e,{depth:t=3,checkOpacity:n=!0,checkVisibilityCSS:r=!0}={}){if(e.checkVisibility)return e.checkVisibility({checkOpacity:n,checkVisibilityCSS:r});let i=e;for(;i&&t>0;){let e=getComputedStyle(i);if(n&&e.opacity===`0`||r&&e.visibility===`hidden`||e.display===`none`)return!1;i=i.parentElement,t--}return!0}function Ct(e,t,n,r){let i=r.x-n.x,a=r.y-n.y,o=i*i+a*a;if(o===0)return 0;let s=((e-n.x)*i+(t-n.y)*a)/o;return Math.max(0,Math.min(1,s))}function k(e,t){return wt(e,e=>e===t)||Tt(e,t)}function wt(e,t){let n;for(n of e.querySelectorAll(`style:not([media])`)??[]){let e;try{e=n.sheet?.cssRules}catch{continue}for(let n of e??[])if(t(n.selectorText))return n}}function Tt(e,t){let n=e.querySelectorAll(`style:not([media])`)??[],r=n?.[n.length-1];if(!r?.sheet)return console.warn(`Media Chrome: No style sheet found on style tag of`,e),{style:{setProperty:()=>{},removeProperty:()=>``,getPropertyValue:()=>``}};let i=r?.sheet.insertRule(`${t}{}`,r.sheet.cssRules.length);return r.sheet.cssRules?.[i]}function A(e,t,n=NaN){let r=e.getAttribute(t);return r==null?n:+r}function j(e,t,n){let r=+n;if(n==null||Number.isNaN(r)){e.hasAttribute(t)&&e.removeAttribute(t);return}A(e,t,void 0)!==r&&e.setAttribute(t,`${r}`)}function M(e,t){return e.hasAttribute(t)}function N(e,t,n){if(n==null){e.hasAttribute(t)&&e.removeAttribute(t);return}M(e,t)!=n&&e.toggleAttribute(t,n)}function P(e,t,n=null){return e.getAttribute(t)??n}function F(e,t,n){if(n==null){e.hasAttribute(t)&&e.removeAttribute(t);return}let r=`${n}`;P(e,t,void 0)!==r&&e.setAttribute(t,r)}var Et=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Dt=(e,t,n)=>(Et(e,t,`read from private field`),n?n.call(e):t.get(e)),Ot=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},kt=(e,t,n,r)=>(Et(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),I;function At(e){return`
    <style>
      :host {
        display: var(--media-control-display, var(--media-gesture-receiver-display, inline-block));
        box-sizing: border-box;
      }
    </style>
  `}var jt=class extends D.HTMLElement{constructor(){if(super(),Ot(this,I,void 0),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=ft(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}}static get observedAttributes(){return[y.MEDIA_CONTROLLER,b.MEDIA_PAUSED]}attributeChangedCallback(e,t,n){var r,i,a,o;e===y.MEDIA_CONTROLLER&&(t&&((i=(r=Dt(this,I))?.unassociateElement)==null||i.call(r,this),kt(this,I,null)),n&&this.isConnected&&(kt(this,I,this.getRootNode()?.getElementById(n)),(o=(a=Dt(this,I))?.associateElement)==null||o.call(a,this)))}connectedCallback(){var e,t;this.tabIndex=-1,this.setAttribute(`aria-hidden`,`true`),kt(this,I,Mt(this)),this.getAttribute(y.MEDIA_CONTROLLER)&&((t=(e=Dt(this,I))?.associateElement)==null||t.call(e,this)),Dt(this,I)&&(Dt(this,I).addEventListener(`pointerdown`,this),Dt(this,I).addEventListener(`click`,this),Dt(this,I).hasAttribute(`tabindex`)||(Dt(this,I).tabIndex=0))}disconnectedCallback(){var e,t,n,r;this.getAttribute(y.MEDIA_CONTROLLER)&&((t=(e=Dt(this,I))?.unassociateElement)==null||t.call(e,this)),(n=Dt(this,I))==null||n.removeEventListener(`pointerdown`,this),(r=Dt(this,I))==null||r.removeEventListener(`click`,this),kt(this,I,null)}handleEvent(e){let t=e.composedPath()?.[0];if([`video`,`media-controller`].includes(t?.localName)){if(e.type===`pointerdown`)this._pointerType=e.pointerType;else if(e.type===`click`){let{clientX:t,clientY:n}=e,{left:r,top:i,width:a,height:o}=this.getBoundingClientRect(),s=t-r,c=n-i;if(s<0||c<0||s>a||c>o||a===0&&o===0)return;let l=this._pointerType||`mouse`;if(this._pointerType=void 0,l===Le.TOUCH){this.handleTap(e);return}if(l===Le.MOUSE||l===Le.PEN){this.handleMouseClick(e);return}}}}get mediaPaused(){return M(this,b.MEDIA_PAUSED)}set mediaPaused(e){N(this,b.MEDIA_PAUSED,e)}handleTap(e){}handleMouseClick(e){let t=this.mediaPaused?v.MEDIA_PLAY_REQUEST:v.MEDIA_PAUSE_REQUEST;this.dispatchEvent(new D.CustomEvent(t,{composed:!0,bubbles:!0}))}};I=new WeakMap,jt.shadowRootOptions={mode:`open`},jt.getTemplateHTML=At;function Mt(e){let t=e.getAttribute(y.MEDIA_CONTROLLER);return t?e.getRootNode()?.getElementById(t):yt(e,`media-controller`)}D.customElements.get(`media-gesture-receiver`)||D.customElements.define(`media-gesture-receiver`,jt);var Nt=jt,Pt=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},L=(e,t,n)=>(Pt(e,t,`read from private field`),n?n.call(e):t.get(e)),R=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Ft=(e,t,n,r)=>(Pt(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),It=(e,t,n)=>(Pt(e,t,`access private method`),n),Lt,Rt,zt,Bt,Vt,Ht,Ut,Wt,Gt,Kt,qt,Jt,Yt,Xt,Zt,Qt,$t,en,tn,nn,z={AUDIO:`audio`,AUTOHIDE:`autohide`,BREAKPOINTS:`breakpoints`,GESTURES_DISABLED:`gesturesdisabled`,KEYBOARD_CONTROL:`keyboardcontrol`,NO_AUTOHIDE:`noautohide`,USER_INACTIVE:`userinactive`,AUTOHIDE_OVER_CONTROLS:`autohideovercontrols`};function rn(e){return`
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

      :host(:not([${z.AUDIO}])) [part~=layer]:not([part~=media-layer]) {
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

      
      :host([${z.AUDIO}]) slot[name=media] {
        display: var(--media-slot-display, none);
      }

      
      :host([${z.AUDIO}]) [part~=layer][part~=gesture-layer] {
        height: 0;
        display: block;
      }

      
      :host(:not([${z.AUDIO}])[${z.GESTURES_DISABLED}]) ::slotted([slot=gestures-chrome]),
          :host(:not([${z.AUDIO}])[${z.GESTURES_DISABLED}]) media-gesture-receiver[slot=gestures-chrome] {
        display: none;
      }

      
      ::slotted(:not([slot=media]):not([slot=poster]):not(media-loading-indicator):not([role=dialog]):not([hidden])) {
        pointer-events: auto;
      }

      :host(:not([${z.AUDIO}])) *[part~=layer][part~=centered-layer] {
        align-items: center;
        justify-content: center;
      }

      :host(:not([${z.AUDIO}])) ::slotted(media-gesture-receiver[slot=gestures-chrome]),
      :host(:not([${z.AUDIO}])) media-gesture-receiver[slot=gestures-chrome] {
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

      
      :host(:not([${z.AUDIO}])) .spacer {
        flex-grow: 1;
      }

      
      :host(:-webkit-full-screen) {
        
        width: 100% !important;
        height: 100% !important;
      }

      
      ::slotted(:not([slot=media]):not([slot=poster]):not([${z.NO_AUTOHIDE}]):not([hidden]):not([role=dialog])) {
        opacity: 1;
        transition: var(--media-control-transition-in, opacity 0.25s);
      }

      
      :host([${z.USER_INACTIVE}]:not([${b.MEDIA_PAUSED}]):not([${b.MEDIA_IS_AIRPLAYING}]):not([${b.MEDIA_IS_CASTING}]):not([${z.AUDIO}])) ::slotted(:not([slot=media]):not([slot=poster]):not([${z.NO_AUTOHIDE}]):not([role=dialog])) {
        opacity: 0;
        transition: var(--media-control-transition-out, opacity 1s);
      }

      :host([${z.USER_INACTIVE}]:not([${z.NO_AUTOHIDE}]):not([${b.MEDIA_PAUSED}]):not([${b.MEDIA_IS_CASTING}]):not([${z.AUDIO}])) ::slotted([slot=media]) {
        cursor: none;
      }

      :host([${z.USER_INACTIVE}][${z.AUTOHIDE_OVER_CONTROLS}]:not([${z.NO_AUTOHIDE}]):not([${b.MEDIA_PAUSED}]):not([${b.MEDIA_IS_CASTING}]):not([${z.AUDIO}])) * {
        --media-cursor: none;
        cursor: none;
      }


      ::slotted(media-control-bar)  {
        align-self: stretch;
      }

      
      :host(:not([${z.AUDIO}])[${b.MEDIA_HAS_PLAYED}]) slot[name=poster] {
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
        <template shadowrootmode="${Nt.shadowRootOptions.mode}">
          ${Nt.getTemplateHTML({})}
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
  `}var an=Object.values(b),on=`sm:384 md:576 lg:768 xl:960`;function sn(e){cn(e.target,e.contentRect.width)}function cn(e,t){if(!e.isConnected)return;let n=ln(e.getAttribute(z.BREAKPOINTS)??on),r=un(n,t),i=!1;if(Object.keys(n).forEach(t=>{if(r.includes(t)){e.hasAttribute(`breakpoint${t}`)||(e.setAttribute(`breakpoint${t}`,``),i=!0);return}e.hasAttribute(`breakpoint${t}`)&&(e.removeAttribute(`breakpoint${t}`),i=!0)}),i){let t=new CustomEvent(Fe.BREAKPOINTS_CHANGE,{detail:r});e.dispatchEvent(t)}e.breakpointsComputed||(e.breakpointsComputed=!0,e.dispatchEvent(new CustomEvent(Fe.BREAKPOINTS_COMPUTED,{bubbles:!0,composed:!0})))}function ln(e){let t=e.split(/\s+/);return Object.fromEntries(t.map(e=>e.split(`:`)))}function un(e,t){return Object.keys(e).filter(n=>t>=parseInt(e[n]))}var dn=class extends D.HTMLElement{constructor(){if(super(),R(this,Gt),R(this,qt),R(this,Yt),R(this,Zt),R(this,$t),R(this,Lt,void 0),R(this,Rt,0),R(this,zt,null),R(this,Bt,null),R(this,Vt,void 0),this.breakpointsComputed=!1,R(this,Ht,e=>{let t=this.media;for(let n of e){if(n.type!==`childList`)continue;let e=n.removedNodes;for(let r of e){if(r.slot!=`media`||n.target!=this)continue;let e=n.previousSibling&&n.previousSibling.previousElementSibling;if(!e||!t)this.mediaUnsetCallback(r);else{let t=e.slot!==`media`;for(;(e=e.previousSibling)!==null;)e.slot==`media`&&(t=!1);t&&this.mediaUnsetCallback(r)}}if(t)for(let e of n.addedNodes)e===t&&this.handleMediaUpdated(t)}}),R(this,Ut,!1),R(this,Wt,e=>{L(this,Ut)||(setTimeout(()=>{sn(e),Ft(this,Ut,!1)},0),Ft(this,Ut,!0))}),R(this,tn,void 0),R(this,nn,()=>{if(!L(this,tn).assignedElements({flatten:!0}).length){L(this,zt)&&this.mediaUnsetCallback(L(this,zt));return}this.handleMediaUpdated(this.media)}),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=ft(this.attributes),t=this.constructor.getTemplateHTML(e);this.shadowRoot.setHTMLUnsafe?this.shadowRoot.setHTMLUnsafe(t):this.shadowRoot.innerHTML=t}Ft(this,Lt,new MutationObserver(L(this,Ht)))}static get observedAttributes(){return[z.AUTOHIDE,z.GESTURES_DISABLED].concat(an).filter(e=>![b.MEDIA_RENDITION_LIST,b.MEDIA_AUDIO_TRACK_LIST,b.MEDIA_CHAPTERS_CUES,b.MEDIA_WIDTH,b.MEDIA_HEIGHT,b.MEDIA_ERROR,b.MEDIA_ERROR_MESSAGE].includes(e))}attributeChangedCallback(e,t,n){e.toLowerCase()==z.AUTOHIDE&&(this.autohide=n)}get media(){let e=this.querySelector(`:scope > [slot=media]`);return e?.nodeName==`SLOT`&&(e=e.assignedElements({flatten:!0})[0]),e}async handleMediaUpdated(e){e&&(Ft(this,zt,e),e.localName.includes(`-`)&&await D.customElements.whenDefined(e.localName),this.mediaSetCallback(e))}connectedCallback(){var e;L(this,Lt).observe(this,{childList:!0,subtree:!0}),ut(this,L(this,Wt));let t=this.getAttribute(z.AUDIO)==null?E(`video player`):E(`audio player`);this.setAttribute(`role`,`region`),this.setAttribute(`aria-label`,t),this.handleMediaUpdated(this.media),this.setAttribute(z.USER_INACTIVE,``),cn(this,this.getBoundingClientRect().width);let n=this.querySelector(`:scope > slot[slot=media]`);n&&(Ft(this,tn,n),L(this,tn).addEventListener(`slotchange`,L(this,nn))),this.addEventListener(`pointerdown`,this),this.addEventListener(`pointermove`,this),this.addEventListener(`pointerup`,this),this.addEventListener(`mouseleave`,this),this.addEventListener(`keyup`,this),(e=D.window)==null||e.addEventListener(`mouseup`,this)}disconnectedCallback(){var e;dt(this,L(this,Wt)),clearTimeout(L(this,Bt)),L(this,Lt).disconnect(),this.media&&this.mediaUnsetCallback(this.media),(e=D.window)==null||e.removeEventListener(`mouseup`,this),this.removeEventListener(`pointerdown`,this),this.removeEventListener(`pointermove`,this),this.removeEventListener(`pointerup`,this),this.removeEventListener(`mouseleave`,this),this.removeEventListener(`keyup`,this),L(this,tn)&&(L(this,tn).removeEventListener(`slotchange`,L(this,nn)),Ft(this,tn,null)),Ft(this,Ut,!1)}mediaSetCallback(e){}mediaUnsetCallback(e){Ft(this,zt,null)}handleEvent(e){switch(e.type){case`pointerdown`:Ft(this,Rt,e.timeStamp);break;case`pointermove`:It(this,Gt,Kt).call(this,e);break;case`pointerup`:It(this,qt,Jt).call(this,e);break;case`mouseleave`:It(this,Yt,Xt).call(this);break;case`mouseup`:this.removeAttribute(z.KEYBOARD_CONTROL);break;case`keyup`:It(this,$t,en).call(this),this.setAttribute(z.KEYBOARD_CONTROL,``)}}set autohide(e){let t=Number(e);Ft(this,Vt,isNaN(t)?0:t)}get autohide(){return(L(this,Vt)===void 0?2:L(this,Vt)).toString()}get breakpoints(){return P(this,z.BREAKPOINTS)}set breakpoints(e){F(this,z.BREAKPOINTS,e)}get audio(){return M(this,z.AUDIO)}set audio(e){N(this,z.AUDIO,e)}get gesturesDisabled(){return M(this,z.GESTURES_DISABLED)}set gesturesDisabled(e){N(this,z.GESTURES_DISABLED,e)}get keyboardControl(){return M(this,z.KEYBOARD_CONTROL)}set keyboardControl(e){N(this,z.KEYBOARD_CONTROL,e)}get noAutohide(){return M(this,z.NO_AUTOHIDE)}set noAutohide(e){N(this,z.NO_AUTOHIDE,e)}get autohideOverControls(){return M(this,z.AUTOHIDE_OVER_CONTROLS)}set autohideOverControls(e){N(this,z.AUTOHIDE_OVER_CONTROLS,e)}get userInteractive(){return M(this,z.USER_INACTIVE)}set userInteractive(e){N(this,z.USER_INACTIVE,e)}};Lt=new WeakMap,Rt=new WeakMap,zt=new WeakMap,Bt=new WeakMap,Vt=new WeakMap,Ht=new WeakMap,Ut=new WeakMap,Wt=new WeakMap,Gt=new WeakSet,Kt=function(e){if(e.pointerType!==`mouse`&&e.timeStamp-L(this,Rt)<250)return;It(this,Zt,Qt).call(this),clearTimeout(L(this,Bt));let t=this.hasAttribute(z.AUTOHIDE_OVER_CONTROLS);([this,this.media].includes(e.target)||t)&&It(this,$t,en).call(this)},qt=new WeakSet,Jt=function(e){if(e.pointerType===`touch`){let t=!this.hasAttribute(z.USER_INACTIVE);[this,this.media].includes(e.target)&&t?It(this,Yt,Xt).call(this):It(this,$t,en).call(this)}else e.composedPath().some(e=>[`media-play-button`,`media-fullscreen-button`].includes(e?.localName))&&It(this,$t,en).call(this)},Yt=new WeakSet,Xt=function(){if(L(this,Vt)<0||this.hasAttribute(z.USER_INACTIVE))return;this.setAttribute(z.USER_INACTIVE,``);let e=new D.CustomEvent(Fe.USER_INACTIVE_CHANGE,{composed:!0,bubbles:!0,detail:!0});this.dispatchEvent(e)},Zt=new WeakSet,Qt=function(){if(!this.hasAttribute(z.USER_INACTIVE))return;this.removeAttribute(z.USER_INACTIVE);let e=new D.CustomEvent(Fe.USER_INACTIVE_CHANGE,{composed:!0,bubbles:!0,detail:!1});this.dispatchEvent(e)},$t=new WeakSet,en=function(){It(this,Zt,Qt).call(this),clearTimeout(L(this,Bt));let e=parseInt(this.autohide);e<0||Ft(this,Bt,setTimeout(()=>{It(this,Yt,Xt).call(this)},e*1e3))},tn=new WeakMap,nn=new WeakMap,dn.shadowRootOptions={mode:`open`},dn.getTemplateHTML=rn,D.customElements.get(`media-container`)||D.customElements.define(`media-container`,dn);var fn=dn,pn=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},B=(e,t,n)=>(pn(e,t,`read from private field`),n?n.call(e):t.get(e)),mn=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},hn=(e,t,n,r)=>(pn(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),gn,_n,vn,yn,bn,xn,Sn=class{constructor(e,t,{defaultValue:n}={defaultValue:void 0}){mn(this,bn),mn(this,gn,void 0),mn(this,_n,void 0),mn(this,vn,void 0),mn(this,yn,new Set),hn(this,gn,e),hn(this,_n,t),hn(this,vn,new Set(n))}[Symbol.iterator](){return B(this,bn,xn).values()}get length(){return B(this,bn,xn).size}get value(){return[...B(this,bn,xn)].join(` `)??``}set value(e){e!==this.value&&(hn(this,yn,new Set),this.add(...e?.split(` `)??[]))}toString(){return this.value}item(e){return[...B(this,bn,xn)][e]}values(){return B(this,bn,xn).values()}forEach(e,t){B(this,bn,xn).forEach(e,t)}add(...e){var t;e.forEach(e=>B(this,yn).add(e)),(this.value!==``||B(this,gn)?.hasAttribute(`${B(this,_n)}`))&&((t=B(this,gn))==null||t.setAttribute(`${B(this,_n)}`,`${this.value}`))}remove(...e){var t;e.forEach(e=>B(this,yn).delete(e)),(t=B(this,gn))==null||t.setAttribute(`${B(this,_n)}`,`${this.value}`)}contains(e){return B(this,bn,xn).has(e)}toggle(e,t){return t===void 0?this.contains(e)?(this.remove(e),!1):(this.add(e),!0):t?(this.add(e),!0):(this.remove(e),!1)}replace(e,t){return this.remove(e),this.add(t),e===t}};gn=new WeakMap,_n=new WeakMap,vn=new WeakMap,yn=new WeakMap,bn=new WeakSet,xn=function(){return B(this,yn).size?B(this,yn):B(this,vn)};var Cn=(e=``)=>e.split(/\s+/),wn=(e=``)=>{let[t,n,r]=e.split(`:`),i=r?decodeURIComponent(r):void 0;return{kind:t===`cc`?S.CAPTIONS:S.SUBTITLES,language:n,label:i}},Tn=(e=``,t={})=>Cn(e).map(e=>{let n=wn(e);return{...t,...n}}),En=e=>e?Array.isArray(e)?e.map(e=>typeof e==`string`?wn(e):e):typeof e==`string`?Tn(e):[e]:[],Dn=({kind:e,label:t,language:n}={kind:`subtitles`})=>t?`${e===`captions`?`cc`:`sb`}:${n}:${encodeURIComponent(t)}`:n,On=(e=[])=>Array.prototype.map.call(e,Dn).join(` `),kn=(e,t)=>n=>n[e]===t,An=e=>{let t=Object.entries(e).map(([e,t])=>kn(e,t));return e=>t.every(t=>t(e))},jn=(e,t=[],n=[])=>{let r=En(n).map(An);Array.from(t).filter(e=>r.some(t=>t(e))).forEach(t=>{t.mode=e})},Mn=(e,t=()=>!0)=>{if(!e?.textTracks)return[];let n=typeof t==`function`?t:An(t);return Array.from(e.textTracks).filter(n)},Nn=e=>!!e.mediaSubtitlesShowing?.length||e.hasAttribute(b.MEDIA_SUBTITLES_SHOWING),Pn=e=>{let{media:t,fullscreenElement:n}=e;try{let e=n&&`requestFullscreen`in n?`requestFullscreen`:n&&`webkitRequestFullScreen`in n?`webkitRequestFullScreen`:void 0;if(e){let t=n[e]?.call(n);if(t instanceof Promise)return t.catch(()=>{})}else t?.webkitEnterFullscreen?t.webkitEnterFullscreen():t?.requestFullscreen&&t.requestFullscreen()}catch(e){console.error(e)}},Fn=`exitFullscreen`in O?`exitFullscreen`:`webkitExitFullscreen`in O?`webkitExitFullscreen`:`webkitCancelFullScreen`in O?`webkitCancelFullScreen`:void 0,In=e=>{let{documentElement:t}=e;if(Fn){let e=(t?.[Fn])?.call(t);if(e instanceof Promise)return e.catch(()=>{})}},Ln=`fullscreenElement`in O?`fullscreenElement`:`webkitFullscreenElement`in O?`webkitFullscreenElement`:void 0,Rn=e=>{let{documentElement:t,media:n}=e,r=t?.[Ln];return!r&&`webkitDisplayingFullscreen`in n&&`webkitPresentationMode`in n&&n.webkitDisplayingFullscreen&&n.webkitPresentationMode===T.FULLSCREEN?n:r},zn=e=>{let{media:t,documentElement:n,fullscreenElement:r=t}=e;if(!t||!n)return!1;let i=Rn(e);if(!i)return!1;if(i===r||i===t)return!0;if(i.localName.includes(`-`)){let e=i.shadowRoot;if(!(Ln in e))return vt(i,r);for(;e?.[Ln];){if(e[Ln]===r)return!0;e=e[Ln]?.shadowRoot}}return!1},Bn=`fullscreenEnabled`in O?`fullscreenEnabled`:`webkitFullscreenEnabled`in O?`webkitFullscreenEnabled`:void 0,Vn=e=>{let{documentElement:t,media:n}=e;return!!t?.[Bn]||n&&`webkitSupportsFullscreen`in n},Hn,Un=()=>{var e;return Hn||(Hn=((e=O)?.createElement)?.call(e,`video`),Hn)},Wn=async(e=Un())=>{if(!e)return!1;let t=e.volume;e.volume=t/2+.1;let n=new AbortController,r=await Promise.race([Gn(e,n.signal),Kn(e,t)]);return n.abort(),r},Gn=(e,t)=>new Promise(n=>{e.addEventListener(`volumechange`,()=>n(!0),{signal:t})}),Kn=async(e,t)=>{for(let n=0;n<10;n++){if(e.volume===t)return!1;await Ue(10)}return e.volume!==t},qn=/.*Version\/.*Safari\/.*/.test(D.navigator.userAgent),Jn=(e=Un())=>D.matchMedia(`(display-mode: standalone)`).matches&&qn?!1:typeof e?.requestPictureInPicture==`function`,Yn=(e=Un())=>Vn({documentElement:O,media:e}),Xn=Yn(),Zn=Jn(),Qn=!!D.WebKitPlaybackTargetAvailabilityEvent,$n=!!D.chrome,er=e=>Mn(e.media,e=>[S.SUBTITLES,S.CAPTIONS].includes(e.kind)).sort((e,t)=>e.kind>=t.kind?1:-1),tr=e=>Mn(e.media,e=>e.mode===Ie.SHOWING&&[S.SUBTITLES,S.CAPTIONS].includes(e.kind)),nr=(e,t)=>{let n=er(e),r=tr(e),i=!!r.length;if(n.length){if(t===!1||i&&t!==!0)jn(Ie.DISABLED,n,r);else if(t===!0||!i&&t!==!1){let t=n[0],{options:i}=e;if(!i?.noSubtitlesLangPref){let e=D.localStorage.getItem(`media-chrome-pref-subtitles-lang`),r=e?[e,...D.navigator.languages]:D.navigator.languages,i=n.filter(e=>r.some(t=>e.language.toLowerCase().startsWith(t.split(`-`)[0]))).sort((e,t)=>r.findIndex(t=>e.language.toLowerCase().startsWith(t.split(`-`)[0]))-r.findIndex(e=>t.language.toLowerCase().startsWith(e.split(`-`)[0])));i[0]&&(t=i[0])}let{language:a,label:o,kind:s}=t;jn(Ie.DISABLED,n,r),jn(Ie.SHOWING,n,[{language:a,label:o,kind:s}])}}},rr=(e,t)=>e===t?!0:e==null||t==null||typeof e!=typeof t?!1:typeof e==`number`&&Number.isNaN(e)&&Number.isNaN(t)?!0:typeof e==`object`?Array.isArray(e)?ir(e,t):Object.entries(e).every(([e,n])=>e in t&&rr(n,t[e])):!1,ir=(e,t)=>{let n=Array.isArray(e),r=Array.isArray(t);return n===r?n||r?e.length===t.length&&e.every((e,n)=>rr(e,t[n])):!0:!1},ar=Object.values(w),or,sr=Wn().then(e=>(or=e,or)),cr=async(...e)=>{await Promise.all(e.filter(e=>e).map(async e=>{if(!(`localName`in e&&e instanceof D.HTMLElement))return;let t=e.localName;if(!t.includes(`-`))return;let n=D.customElements.get(t);n&&e instanceof n||(await D.customElements.whenDefined(t),D.customElements.upgrade(e))}))},lr=new D.DOMParser,ur=e=>e&&(lr.parseFromString(e,`text/html`).body.textContent||e),dr={mediaError:{get(e,t){let{media:n}=e;if(t?.type!==`playing`)return n?.error},mediaEvents:[`emptied`,`error`,`playing`]},mediaErrorCode:{get(e,t){let{media:n}=e;if(t?.type!==`playing`)return n?.error?.code},mediaEvents:[`emptied`,`error`,`playing`]},mediaErrorMessage:{get(e,t){let{media:n}=e;if(t?.type!==`playing`)return n?.error?.message??``},mediaEvents:[`emptied`,`error`,`playing`]},mediaWidth:{get(e){let{media:t}=e;return t?.videoWidth??0},mediaEvents:[`resize`]},mediaHeight:{get(e){let{media:t}=e;return t?.videoHeight??0},mediaEvents:[`resize`]},mediaPaused:{get(e){let{media:t}=e;return t?.paused??!0},set(e,t){var n;let{media:r}=t;r&&(e?r.pause():(n=r.play())==null||n.catch(()=>{}))},mediaEvents:[`play`,`playing`,`pause`,`emptied`]},mediaHasPlayed:{get(e,t){let{media:n}=e;return n?t?t.type===`playing`:!n.paused:!1},mediaEvents:[`playing`,`emptied`]},mediaEnded:{get(e){let{media:t}=e;return t?.ended??!1},mediaEvents:[`seeked`,`ended`,`emptied`]},mediaPlaybackRate:{get(e){let{media:t}=e;return t?.playbackRate??1},set(e,t){let{media:n}=t;n&&Number.isFinite(+e)&&(n.playbackRate=+e)},mediaEvents:[`ratechange`,`loadstart`]},mediaMuted:{get(e){let{media:t}=e;return t?.muted??!1},set(e,t){let{media:n,options:{noMutedPref:r}={}}=t;if(n){n.muted=e;try{let t=D.localStorage.getItem(`media-chrome-pref-muted`)!==null,i=n.hasAttribute(`muted`);if(r){t&&D.localStorage.removeItem(`media-chrome-pref-muted`);return}if(i&&!t)return;D.localStorage.setItem(`media-chrome-pref-muted`,e?`true`:`false`)}catch(e){console.debug(`Error setting muted pref`,e)}}},mediaEvents:[`volumechange`],stateOwnersUpdateHandlers:[(e,t)=>{let{options:{noMutedPref:n}}=t,{media:r}=t;if(!(!r||r.muted||n))try{let n=D.localStorage.getItem(`media-chrome-pref-muted`)===`true`;dr.mediaMuted.set(n,t),e(n)}catch(e){console.debug(`Error getting muted pref`,e)}}]},mediaLoop:{get(e){let{media:t}=e;return t?.loop},set(e,t){let{media:n}=t;n&&(n.loop=e)},mediaEvents:[`medialooprequest`]},mediaVolume:{get(e){let{media:t}=e;return t?.volume??1},set(e,t){let{media:n,options:{noVolumePref:r}={}}=t;if(n){try{e==null?D.localStorage.removeItem(`media-chrome-pref-volume`):!n.hasAttribute(`muted`)&&!r&&D.localStorage.setItem(`media-chrome-pref-volume`,e.toString())}catch(e){console.debug(`Error setting volume pref`,e)}Number.isFinite(+e)&&(n.volume=+e)}},mediaEvents:[`volumechange`],stateOwnersUpdateHandlers:[(e,t)=>{let{options:{noVolumePref:n}}=t;if(!n)try{let{media:n}=t;if(!n)return;let r=D.localStorage.getItem(`media-chrome-pref-volume`);if(r==null)return;dr.mediaVolume.set(+r,t),e(+r)}catch(e){console.debug(`Error getting volume pref`,e)}}]},mediaVolumeLevel:{get(e){let{media:t}=e;return t?.volume===void 0?`high`:t.muted||t.volume===0?`off`:t.volume<.5?`low`:t.volume<.75?`medium`:`high`},mediaEvents:[`volumechange`]},mediaCurrentTime:{get(e){let{media:t}=e;return t?.currentTime??0},set(e,t){let{media:n}=t;n&&He(e)&&(n.currentTime=e)},mediaEvents:[`timeupdate`,`loadedmetadata`,`seeking`]},mediaDuration:{get(e){let{media:t,options:{defaultDuration:n}={}}=e;return n&&(!t||!t.duration||Number.isNaN(t.duration)||!Number.isFinite(t.duration))?n:Number.isFinite(t?.duration)?t.duration:NaN},mediaEvents:[`durationchange`,`loadedmetadata`,`emptied`]},mediaLoading:{get(e){let{media:t}=e;return t?.readyState<3},mediaEvents:[`waiting`,`playing`,`emptied`]},mediaSeekable:{get(e){let{media:t}=e;if(!t?.seekable?.length)return;let n=t.seekable.start(0),r=t.seekable.end(t.seekable.length-1);if(n||r)return[Number(n.toFixed(3)),Number(r.toFixed(3))]},mediaEvents:[`loadedmetadata`,`emptied`,`progress`,`seekablechange`]},mediaBuffered:{get(e){let{media:t}=e,n=t?.buffered??[];return Array.from(n).map((e,t)=>[Number(n.start(t).toFixed(3)),Number(n.end(t).toFixed(3))])},mediaEvents:[`progress`,`emptied`]},mediaStreamType:{get(e){let{media:t,options:{defaultStreamType:n}={}}=e,r=[w.LIVE,w.ON_DEMAND].includes(n)?n:void 0;if(!t)return r;let{streamType:i}=t;if(ar.includes(i))return i===w.UNKNOWN?r:i;let a=t.duration;return a===1/0?w.LIVE:Number.isFinite(a)?w.ON_DEMAND:r},mediaEvents:[`emptied`,`durationchange`,`loadedmetadata`,`streamtypechange`]},mediaTargetLiveWindow:{get(e){let{media:t}=e;if(!t)return NaN;let{targetLiveWindow:n}=t,r=dr.mediaStreamType.get(e);return(n==null||Number.isNaN(n))&&r===w.LIVE?0:n},mediaEvents:[`emptied`,`durationchange`,`loadedmetadata`,`streamtypechange`,`targetlivewindowchange`]},mediaTimeIsLive:{get(e){let{media:t,options:{liveEdgeOffset:n=10}={}}=e;if(!t)return!1;if(typeof t.liveEdgeStart==`number`)return!Number.isNaN(t.liveEdgeStart)&&t.currentTime>=t.liveEdgeStart;if(dr.mediaStreamType.get(e)!==w.LIVE)return!1;let r=t.seekable;if(!r)return!0;if(!r.length)return!1;let i=r.end(r.length-1)-n;return t.currentTime>=i},mediaEvents:[`playing`,`timeupdate`,`progress`,`waiting`,`emptied`]},mediaSubtitlesList:{get(e){return er(e).map(({kind:e,label:t,language:n})=>({kind:e,label:t,language:n}))},mediaEvents:[`loadstart`],textTracksEvents:[`addtrack`,`removetrack`]},mediaSubtitlesShowing:{get(e){return tr(e).map(({kind:e,label:t,language:n})=>({kind:e,label:t,language:n}))},mediaEvents:[`loadstart`],textTracksEvents:[`addtrack`,`removetrack`,`change`],stateOwnersUpdateHandlers:[(e,t)=>{var n,r;let{media:i,options:a}=t;if(!i)return;let o=e=>{a.defaultSubtitles&&(!e||[S.CAPTIONS,S.SUBTITLES].includes(e?.track?.kind))&&nr(t,!0)};return i.addEventListener(`loadstart`,o),(n=i.textTracks)==null||n.addEventListener(`addtrack`,o),(r=i.textTracks)==null||r.addEventListener(`removetrack`,o),()=>{var e,t;i.removeEventListener(`loadstart`,o),(e=i.textTracks)==null||e.removeEventListener(`addtrack`,o),(t=i.textTracks)==null||t.removeEventListener(`removetrack`,o)}}]},mediaChaptersCues:{get(e){let{media:t}=e;if(!t)return[];let[n]=Mn(t,{kind:S.CHAPTERS});return Array.from(n?.cues??[]).map(({text:e,startTime:t,endTime:n})=>({text:ur(e),startTime:t,endTime:n}))},mediaEvents:[`loadstart`,`loadedmetadata`],textTracksEvents:[`addtrack`,`removetrack`,`change`],stateOwnersUpdateHandlers:[(e,t)=>{let{media:n}=t;if(!n)return;let r=n.querySelector(`track[kind="chapters"][default][src]`),i=n.shadowRoot?.querySelector(`:is(video,audio) > track[kind="chapters"][default][src]`);return r?.addEventListener(`load`,e),i?.addEventListener(`load`,e),()=>{r?.removeEventListener(`load`,e),i?.removeEventListener(`load`,e)}}]},mediaIsPip:{get(e){let{media:t,documentElement:n}=e;if(!t||!n||!n.pictureInPictureElement)return!1;if(n.pictureInPictureElement===t)return!0;if(n.pictureInPictureElement instanceof HTMLMediaElement)return t.localName?.includes(`-`)?vt(t,n.pictureInPictureElement):!1;if(n.pictureInPictureElement.localName.includes(`-`)){let e=n.pictureInPictureElement.shadowRoot;for(;e?.pictureInPictureElement;){if(e.pictureInPictureElement===t)return!0;e=e.pictureInPictureElement?.shadowRoot}}return!1},set(e,t){let{media:n}=t;if(n){if(e){if(!O.pictureInPictureEnabled){console.warn(`MediaChrome: Picture-in-picture is not enabled`);return}if(!n.requestPictureInPicture){console.warn(`MediaChrome: The current media does not support picture-in-picture`);return}let e=()=>{console.warn(`MediaChrome: The media is not ready for picture-in-picture. It must have a readyState > 0.`)};n.requestPictureInPicture().catch(t=>{if(t.code===11){if(!n.src){console.warn(`MediaChrome: The media is not ready for picture-in-picture. It must have a src set.`);return}if(n.readyState===0&&n.preload===`none`){let t=()=>{n.removeEventListener(`loadedmetadata`,r),n.preload=`none`},r=()=>{n.requestPictureInPicture().catch(e),t()};n.addEventListener(`loadedmetadata`,r),n.preload=`metadata`,setTimeout(()=>{n.readyState===0&&e(),t()},1e3)}else throw t}else throw t})}else O.pictureInPictureElement&&O.exitPictureInPicture()}},mediaEvents:[`enterpictureinpicture`,`leavepictureinpicture`]},mediaRenditionList:{get(e){let{media:t}=e;return[...t?.videoRenditions??[]].map(e=>({...e}))},mediaEvents:[`emptied`,`loadstart`],videoRenditionsEvents:[`addrendition`,`removerendition`]},mediaRenditionSelected:{get(e){let{media:t}=e;return t?.videoRenditions?.[t.videoRenditions?.selectedIndex]?.id},set(e,t){let{media:n}=t;if(!n?.videoRenditions){console.warn(`MediaController: Rendition selection not supported by this media.`);return}let r=e,i=Array.prototype.findIndex.call(n.videoRenditions,e=>e.id==r);n.videoRenditions.selectedIndex!=i&&(n.videoRenditions.selectedIndex=i)},mediaEvents:[`emptied`],videoRenditionsEvents:[`addrendition`,`removerendition`,`change`]},mediaAudioTrackList:{get(e){let{media:t}=e;return[...t?.audioTracks??[]]},mediaEvents:[`emptied`,`loadstart`],audioTracksEvents:[`addtrack`,`removetrack`]},mediaAudioTrackEnabled:{get(e){let{media:t}=e;return[...t?.audioTracks??[]].find(e=>e.enabled)?.id},set(e,t){let{media:n}=t;if(!n?.audioTracks){console.warn(`MediaChrome: Audio track selection not supported by this media.`);return}let r=e;for(let e of n.audioTracks)e.enabled=r==e.id},mediaEvents:[`emptied`],audioTracksEvents:[`addtrack`,`removetrack`,`change`]},mediaIsFullscreen:{get(e){return zn(e)},set(e,t,n){var r;e?(Pn(t),n.detail&&!t.media?.inert&&((r=t.media)==null||r.focus())):In(t)},rootEvents:[`fullscreenchange`,`webkitfullscreenchange`],mediaEvents:[`webkitbeginfullscreen`,`webkitendfullscreen`,`webkitpresentationmodechanged`]},mediaIsCasting:{get(e){let{media:t}=e;return!t?.remote||t.remote?.state===`disconnected`?!1:t.remote.state===`connected`},set(e,t){let{media:n}=t;if(n&&!(e&&n.remote?.state!==`disconnected`)&&(e||n.remote?.state===`connected`)){if(typeof n.remote.prompt!=`function`){console.warn(`MediaChrome: Casting is not supported in this environment`);return}n.remote.prompt().catch(()=>{})}},remoteEvents:[`connect`,`connecting`,`disconnect`]},mediaIsAirplaying:{get(){return!1},set(e,t){let{media:n}=t;if(n){if(!(n.webkitShowPlaybackTargetPicker&&D.WebKitPlaybackTargetAvailabilityEvent)){console.error(`MediaChrome: received a request to select AirPlay but AirPlay is not supported in this environment`);return}n.webkitShowPlaybackTargetPicker()}},mediaEvents:[`webkitcurrentplaybacktargetiswirelesschanged`]},mediaFullscreenUnavailable:{get(e){let{media:t}=e;if(!Xn||!Yn(t))return C.UNSUPPORTED}},mediaPipUnavailable:{get(e){let{media:t}=e;if(!Zn||!Jn(t))return C.UNSUPPORTED;if(t?.disablePictureInPicture)return C.UNAVAILABLE}},mediaVolumeUnavailable:{get(e){let{media:t}=e;if(or===!1||t?.volume==null)return C.UNSUPPORTED},stateOwnersUpdateHandlers:[e=>{or??sr.then(t=>e(t?void 0:C.UNSUPPORTED))}]},mediaCastUnavailable:{get(e,{availability:t=`not-available`}={}){let{media:n}=e;if(!$n||!n?.remote?.state)return C.UNSUPPORTED;if(t!=null&&t!==`available`)return C.UNAVAILABLE},stateOwnersUpdateHandlers:[(e,t)=>{var n;let{media:r}=t;if(r)return r.disableRemotePlayback||r.hasAttribute(`disableremoteplayback`)||(n=r?.remote)==null||n.watchAvailability(t=>{e({availability:t?`available`:`not-available`})}).catch(t=>{t.name===`NotSupportedError`?e({availability:null}):e({availability:`not-available`})}),()=>{var e;(e=r?.remote)==null||e.cancelWatchAvailability().catch(()=>{})}}]},mediaAirplayUnavailable:{get(e,t){if(!Qn)return C.UNSUPPORTED;if(t?.availability===`not-available`)return C.UNAVAILABLE},mediaEvents:[`webkitplaybacktargetavailabilitychanged`],stateOwnersUpdateHandlers:[(e,t)=>{var n;let{media:r}=t;if(r)return r.disableRemotePlayback||r.hasAttribute(`disableremoteplayback`)||(n=r?.remote)==null||n.watchAvailability(t=>{e({availability:t?`available`:`not-available`})}).catch(t=>{t.name===`NotSupportedError`?e({availability:null}):e({availability:`not-available`})}),()=>{var e;(e=r?.remote)==null||e.cancelWatchAvailability().catch(()=>{})}}]},mediaRenditionUnavailable:{get(e){let{media:t}=e;if(!t?.videoRenditions)return C.UNSUPPORTED;if(!t.videoRenditions?.length)return C.UNAVAILABLE},mediaEvents:[`emptied`,`loadstart`],videoRenditionsEvents:[`addrendition`,`removerendition`]},mediaAudioTrackUnavailable:{get(e){let{media:t}=e;if(!t?.audioTracks)return C.UNSUPPORTED;if((t.audioTracks?.length??0)<=1)return C.UNAVAILABLE},mediaEvents:[`emptied`,`loadstart`],audioTracksEvents:[`addtrack`,`removetrack`]},mediaLang:{get(e){let{options:{mediaLang:t}={}}=e;return t??`en`}}},fr={[v.MEDIA_PREVIEW_REQUEST](e,t,{detail:n}){let{media:r}=t,i=n??void 0,a,o;if(r&&i!=null){let[e]=Mn(r,{kind:S.METADATA,label:`thumbnails`}),t=Array.prototype.find.call(e?.cues??[],(e,t,n)=>t===0?e.endTime>i:t===n.length-1?e.startTime<=i:e.startTime<=i&&e.endTime>i);if(t){let e=/'^(?:[a-z]+:)?\/\//i.test(t.text)?void 0:r?.querySelector(`track[label="thumbnails"]`)?.src,n=new URL(t.text,e);o=new URLSearchParams(n.hash).get(`#xywh`).split(`,`).map(e=>+e),a=n.href}}let s=e.mediaDuration.get(t),c=e.mediaChaptersCues.get(t).find((e,t,n)=>t===n.length-1&&s===e.endTime?e.startTime<=i&&e.endTime>=i:e.startTime<=i&&e.endTime>i)?.text;return n!=null&&c==null&&(c=``),{mediaPreviewTime:i,mediaPreviewImage:a,mediaPreviewCoords:o,mediaPreviewChapter:c}},[v.MEDIA_PAUSE_REQUEST](e,t){e.mediaPaused.set(!0,t)},[v.MEDIA_PLAY_REQUEST](e,t){let n=e.mediaStreamType.get(t)===w.LIVE,r=!t.options?.noAutoSeekToLive,i=e.mediaTargetLiveWindow.get(t)>0;if(n&&r&&!i){let n=e.mediaSeekable.get(t)?.[1];if(n){let r=n-(t.options?.seekToLiveOffset??0);e.mediaCurrentTime.set(r,t)}}e.mediaPaused.set(!1,t)},[v.MEDIA_PLAYBACK_RATE_REQUEST](e,t,{detail:n}){let r=n;e.mediaPlaybackRate.set(r,t)},[v.MEDIA_MUTE_REQUEST](e,t){e.mediaMuted.set(!0,t)},[v.MEDIA_UNMUTE_REQUEST](e,t){e.mediaVolume.get(t)||e.mediaVolume.set(.25,t),e.mediaMuted.set(!1,t)},[v.MEDIA_LOOP_REQUEST](e,t,{detail:n}){let r=!!n;return e.mediaLoop.set(r,t),{mediaLoop:r}},[v.MEDIA_VOLUME_REQUEST](e,t,{detail:n}){let r=n;r&&e.mediaMuted.get(t)&&e.mediaMuted.set(!1,t),e.mediaVolume.set(r,t)},[v.MEDIA_SEEK_REQUEST](e,t,{detail:n}){let r=n;e.mediaCurrentTime.set(r,t)},[v.MEDIA_SEEK_TO_LIVE_REQUEST](e,t){let n=e.mediaSeekable.get(t)?.[1];if(Number.isNaN(Number(n)))return;let r=n-(t.options?.seekToLiveOffset??0);e.mediaCurrentTime.set(r,t)},[v.MEDIA_SHOW_SUBTITLES_REQUEST](e,t,{detail:n}){let{options:r}=t,i=er(t),a=En(n),o=a[0]?.language;o&&!r.noSubtitlesLangPref&&D.localStorage.setItem(`media-chrome-pref-subtitles-lang`,o),jn(Ie.SHOWING,i,a)},[v.MEDIA_DISABLE_SUBTITLES_REQUEST](e,t,{detail:n}){let r=er(t),i=n??[];jn(Ie.DISABLED,r,i)},[v.MEDIA_TOGGLE_SUBTITLES_REQUEST](e,t,{detail:n}){nr(t,n)},[v.MEDIA_RENDITION_REQUEST](e,t,{detail:n}){let r=n;e.mediaRenditionSelected.set(r,t)},[v.MEDIA_AUDIO_TRACK_REQUEST](e,t,{detail:n}){let r=n;e.mediaAudioTrackEnabled.set(r,t)},[v.MEDIA_ENTER_PIP_REQUEST](e,t){e.mediaIsFullscreen.get(t)&&e.mediaIsFullscreen.set(!1,t),e.mediaIsPip.set(!0,t)},[v.MEDIA_EXIT_PIP_REQUEST](e,t){e.mediaIsPip.set(!1,t)},[v.MEDIA_ENTER_FULLSCREEN_REQUEST](e,t,n){e.mediaIsPip.get(t)&&e.mediaIsPip.set(!1,t),e.mediaIsFullscreen.set(!0,t,n)},[v.MEDIA_EXIT_FULLSCREEN_REQUEST](e,t){e.mediaIsFullscreen.set(!1,t)},[v.MEDIA_ENTER_CAST_REQUEST](e,t){e.mediaIsFullscreen.get(t)&&e.mediaIsFullscreen.set(!1,t),e.mediaIsCasting.set(!0,t)},[v.MEDIA_EXIT_CAST_REQUEST](e,t){e.mediaIsCasting.set(!1,t)},[v.MEDIA_AIRPLAY_REQUEST](e,t){e.mediaIsAirplaying.set(!0,t)}},pr=({media:e,fullscreenElement:t,documentElement:n,stateMediator:r=dr,requestMap:i=fr,options:a={},monitorStateOwnersOnlyWithSubscriptions:o=!0})=>{let s=[],c={options:{...a}},l=Object.freeze({mediaPreviewTime:void 0,mediaPreviewImage:void 0,mediaPreviewCoords:void 0,mediaPreviewChapter:void 0}),u=e=>{e!=null&&(rr(e,l)||(l=Object.freeze({...l,...e}),s.forEach(e=>e(l))))},d=()=>{let e=Object.entries(r).reduce((e,[t,{get:n}])=>(e[t]=n(c),e),{});u(e)},f={},p,m=async(e,t)=>{let n=!!p;if(p={...c,...p??{},...e},n)return;await cr(...Object.values(e));let i=s.length>0&&t===0&&o,a=c.media!==p.media,l=c.media?.textTracks!==p.media?.textTracks,m=c.media?.videoRenditions!==p.media?.videoRenditions,h=c.media?.audioTracks!==p.media?.audioTracks,ee=c.media?.remote!==p.media?.remote,te=c.documentElement!==p.documentElement,ne=!!c.media&&(a||i),re=!!c.media?.textTracks&&(l||i),ie=!!c.media?.videoRenditions&&(m||i),ae=!!c.media?.audioTracks&&(h||i),oe=!!c.media?.remote&&(ee||i),se=!!c.documentElement&&(te||i),ce=ne||re||ie||ae||oe||se,le=s.length===0&&t===1&&o,ue=!!p.media&&(a||le),de=!!p.media?.textTracks&&(l||le),fe=!!p.media?.videoRenditions&&(m||le),pe=!!p.media?.audioTracks&&(h||le),me=!!p.media?.remote&&(ee||le),he=!!p.documentElement&&(te||le),ge=ue||de||fe||pe||me||he;if(!(ce||ge)){Object.entries(p).forEach(([e,t])=>{c[e]=t}),d(),p=void 0;return}Object.entries(r).forEach(([e,{get:t,mediaEvents:n=[],textTracksEvents:r=[],videoRenditionsEvents:i=[],audioTracksEvents:a=[],remoteEvents:o=[],rootEvents:s=[],stateOwnersUpdateHandlers:l=[]}])=>{f[e]||(f[e]={});let d=n=>{let r=t(c,n);u({[e]:r})},m;m=f[e].mediaEvents,n.forEach(t=>{m&&ne&&(c.media.removeEventListener(t,m),f[e].mediaEvents=void 0),ue&&(p.media.addEventListener(t,d),f[e].mediaEvents=d)}),m=f[e].textTracksEvents,r.forEach(t=>{var n,r;m&&re&&((n=c.media.textTracks)==null||n.removeEventListener(t,m),f[e].textTracksEvents=void 0),de&&((r=p.media.textTracks)==null||r.addEventListener(t,d),f[e].textTracksEvents=d)}),m=f[e].videoRenditionsEvents,i.forEach(t=>{var n,r;m&&ie&&((n=c.media.videoRenditions)==null||n.removeEventListener(t,m),f[e].videoRenditionsEvents=void 0),fe&&((r=p.media.videoRenditions)==null||r.addEventListener(t,d),f[e].videoRenditionsEvents=d)}),m=f[e].audioTracksEvents,a.forEach(t=>{var n,r;m&&ae&&((n=c.media.audioTracks)==null||n.removeEventListener(t,m),f[e].audioTracksEvents=void 0),pe&&((r=p.media.audioTracks)==null||r.addEventListener(t,d),f[e].audioTracksEvents=d)}),m=f[e].remoteEvents,o.forEach(t=>{var n,r;m&&oe&&((n=c.media.remote)==null||n.removeEventListener(t,m),f[e].remoteEvents=void 0),me&&((r=p.media.remote)==null||r.addEventListener(t,d),f[e].remoteEvents=d)}),m=f[e].rootEvents,s.forEach(t=>{m&&se&&(c.documentElement.removeEventListener(t,m),f[e].rootEvents=void 0),he&&(p.documentElement.addEventListener(t,d),f[e].rootEvents=d)});let h=f[e].stateOwnersUpdateHandlers;if(h&&ce&&(Array.isArray(h)?h:[h]).forEach(e=>{typeof e==`function`&&e()}),ge){let t=l.map(e=>e(d,p)).filter(e=>typeof e==`function`);f[e].stateOwnersUpdateHandlers=t.length===1?t[0]:t}else ce&&(f[e].stateOwnersUpdateHandlers=void 0)}),Object.entries(p).forEach(([e,t])=>{c[e]=t}),d(),p=void 0};return m({media:e,fullscreenElement:t,documentElement:n,options:a}),{dispatch(e){let{type:t,detail:n}=e;if(i[t]&&l.mediaErrorCode==null){u(i[t](r,c,e));return}t===`mediaelementchangerequest`?m({media:n}):t===`fullscreenelementchangerequest`?m({fullscreenElement:n}):t===`documentelementchangerequest`?m({documentElement:n}):t===`optionschangerequest`&&(Object.entries(n??{}).forEach(([e,t])=>{c.options[e]=t}),d())},getState(){return l},subscribe(e){return m({},s.length+1),s.push(e),e(l),()=>{let t=s.indexOf(e);t>=0&&(m({},s.length-1),s.splice(t,1))}}}},mr=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},V=(e,t,n)=>(mr(e,t,`read from private field`),n?n.call(e):t.get(e)),hr=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},gr=(e,t,n,r)=>(mr(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),_r=(e,t,n)=>(mr(e,t,`access private method`),n),vr,yr,H,br,xr,Sr,Cr,wr,Tr,Er,Dr,Or,kr,Ar,jr,Mr=[`ArrowLeft`,`ArrowRight`,`ArrowUp`,`ArrowDown`,`Enter`,` `,`f`,`m`,`k`,`c`,`l`,`j`,`>`,`<`,`p`],Nr=10,Pr=.025,Fr=.25,Ir=.25,Lr=2,U={DEFAULT_SUBTITLES:`defaultsubtitles`,DEFAULT_STREAM_TYPE:`defaultstreamtype`,DEFAULT_DURATION:`defaultduration`,FULLSCREEN_ELEMENT:`fullscreenelement`,HOTKEYS:`hotkeys`,KEYBOARD_BACKWARD_SEEK_OFFSET:`keyboardbackwardseekoffset`,KEYBOARD_FORWARD_SEEK_OFFSET:`keyboardforwardseekoffset`,KEYBOARD_DOWN_VOLUME_STEP:`keyboarddownvolumestep`,KEYBOARD_UP_VOLUME_STEP:`keyboardupvolumestep`,KEYS_USED:`keysused`,LANG:`lang`,LOOP:`loop`,LIVE_EDGE_OFFSET:`liveedgeoffset`,NO_AUTO_SEEK_TO_LIVE:`noautoseektolive`,NO_DEFAULT_STORE:`nodefaultstore`,NO_HOTKEYS:`nohotkeys`,NO_MUTED_PREF:`nomutedpref`,NO_SUBTITLES_LANG_PREF:`nosubtitleslangpref`,NO_VOLUME_PREF:`novolumepref`,SEEK_TO_LIVE_OFFSET:`seektoliveoffset`},Rr=class extends dn{constructor(){super(),hr(this,Tr),hr(this,Or),hr(this,Ar),this.mediaStateReceivers=[],this.associatedElementSubscriptions=new Map,hr(this,vr,new Sn(this,U.HOTKEYS)),hr(this,yr,void 0),hr(this,H,void 0),hr(this,br,null),hr(this,xr,void 0),hr(this,Sr,void 0),hr(this,Cr,e=>{var t;(t=V(this,H))==null||t.dispatch(e)}),hr(this,wr,void 0),hr(this,Dr,e=>{let{key:t,shiftKey:n}=e;if(!(n&&(t===`/`||t===`?`)||Mr.includes(t))){this.removeEventListener(`keyup`,V(this,Dr));return}this.keyboardShortcutHandler(e)}),this.associateElement(this);let e={};gr(this,xr,t=>{Object.entries(t).forEach(([t,n])=>{if(t in e&&e[t]===n)return;this.propagateMediaState(t,n);let r=t.toLowerCase(),i=new D.CustomEvent(x[r],{composed:!0,detail:n});this.dispatchEvent(i)}),e=t})}static get observedAttributes(){return super.observedAttributes.concat(U.NO_HOTKEYS,U.HOTKEYS,U.DEFAULT_STREAM_TYPE,U.DEFAULT_SUBTITLES,U.DEFAULT_DURATION,U.NO_MUTED_PREF,U.NO_VOLUME_PREF,U.LANG,U.LOOP,U.LIVE_EDGE_OFFSET,U.SEEK_TO_LIVE_OFFSET,U.NO_AUTO_SEEK_TO_LIVE)}get mediaStore(){return V(this,H)}set mediaStore(e){var t;if(V(this,H)&&((t=V(this,Sr))==null||t.call(this),gr(this,Sr,void 0)),gr(this,H,e),!V(this,H)&&!this.hasAttribute(U.NO_DEFAULT_STORE)){_r(this,Tr,Er).call(this);return}gr(this,Sr,V(this,H)?.subscribe(V(this,xr)))}get fullscreenElement(){return V(this,yr)??this}set fullscreenElement(e){var t;this.hasAttribute(U.FULLSCREEN_ELEMENT)&&this.removeAttribute(U.FULLSCREEN_ELEMENT),gr(this,yr,e),(t=V(this,H))==null||t.dispatch({type:`fullscreenelementchangerequest`,detail:this.fullscreenElement})}get defaultSubtitles(){return M(this,U.DEFAULT_SUBTITLES)}set defaultSubtitles(e){N(this,U.DEFAULT_SUBTITLES,e)}get defaultStreamType(){return P(this,U.DEFAULT_STREAM_TYPE)}set defaultStreamType(e){F(this,U.DEFAULT_STREAM_TYPE,e)}get defaultDuration(){return A(this,U.DEFAULT_DURATION)}set defaultDuration(e){j(this,U.DEFAULT_DURATION,e)}get noHotkeys(){return M(this,U.NO_HOTKEYS)}set noHotkeys(e){N(this,U.NO_HOTKEYS,e)}get keysUsed(){return P(this,U.KEYS_USED)}set keysUsed(e){F(this,U.KEYS_USED,e)}get liveEdgeOffset(){return A(this,U.LIVE_EDGE_OFFSET)}set liveEdgeOffset(e){j(this,U.LIVE_EDGE_OFFSET,e)}get noAutoSeekToLive(){return M(this,U.NO_AUTO_SEEK_TO_LIVE)}set noAutoSeekToLive(e){N(this,U.NO_AUTO_SEEK_TO_LIVE,e)}get noVolumePref(){return M(this,U.NO_VOLUME_PREF)}set noVolumePref(e){N(this,U.NO_VOLUME_PREF,e)}get noMutedPref(){return M(this,U.NO_MUTED_PREF)}set noMutedPref(e){N(this,U.NO_MUTED_PREF,e)}get noSubtitlesLangPref(){return M(this,U.NO_SUBTITLES_LANG_PREF)}set noSubtitlesLangPref(e){N(this,U.NO_SUBTITLES_LANG_PREF,e)}get noDefaultStore(){return M(this,U.NO_DEFAULT_STORE)}set noDefaultStore(e){N(this,U.NO_DEFAULT_STORE,e)}get resolvedLang(){return Je()}attributeChangedCallback(e,t,n){var r,i,a,o,s,c,l,u,d,f;if(super.attributeChangedCallback(e,t,n),e===U.NO_HOTKEYS)n!==t&&n===``?(this.hasAttribute(U.HOTKEYS)&&console.warn("Media Chrome: Both `hotkeys` and `nohotkeys` have been set. All hotkeys will be disabled."),this.disableHotkeys()):n!==t&&n===null&&this.enableHotkeys();else if(e===U.HOTKEYS)V(this,vr).value=n;else if(e===U.DEFAULT_SUBTITLES&&n!==t)(r=V(this,H))==null||r.dispatch({type:`optionschangerequest`,detail:{defaultSubtitles:this.hasAttribute(U.DEFAULT_SUBTITLES)}});else if(e===U.DEFAULT_STREAM_TYPE)(i=V(this,H))==null||i.dispatch({type:`optionschangerequest`,detail:{defaultStreamType:this.getAttribute(U.DEFAULT_STREAM_TYPE)??void 0}});else if(e===U.LIVE_EDGE_OFFSET&&n!==t)(a=V(this,H))==null||a.dispatch({type:`optionschangerequest`,detail:{liveEdgeOffset:this.hasAttribute(U.LIVE_EDGE_OFFSET)?+this.getAttribute(U.LIVE_EDGE_OFFSET):void 0,seekToLiveOffset:this.hasAttribute(U.SEEK_TO_LIVE_OFFSET)?+this.getAttribute(U.SEEK_TO_LIVE_OFFSET):this.hasAttribute(U.LIVE_EDGE_OFFSET)?+this.getAttribute(U.LIVE_EDGE_OFFSET):void 0}});else if(e===U.SEEK_TO_LIVE_OFFSET&&n!==t)(o=V(this,H))==null||o.dispatch({type:`optionschangerequest`,detail:{seekToLiveOffset:this.hasAttribute(U.SEEK_TO_LIVE_OFFSET)?+this.getAttribute(U.SEEK_TO_LIVE_OFFSET):this.hasAttribute(U.LIVE_EDGE_OFFSET)?+this.getAttribute(U.LIVE_EDGE_OFFSET):void 0}});else if(e===U.NO_AUTO_SEEK_TO_LIVE)(s=V(this,H))==null||s.dispatch({type:`optionschangerequest`,detail:{noAutoSeekToLive:this.hasAttribute(U.NO_AUTO_SEEK_TO_LIVE)}});else if(e===U.FULLSCREEN_ELEMENT){let e=n?this.getRootNode()?.getElementById(n):void 0;gr(this,yr,e),(c=V(this,H))==null||c.dispatch({type:`fullscreenelementchangerequest`,detail:this.fullscreenElement})}else e===U.LANG&&n!==t?(Ke(n),(l=V(this,H))==null||l.dispatch({type:`optionschangerequest`,detail:{mediaLang:n}})):e===U.LOOP&&n!==t?(u=V(this,H))==null||u.dispatch({type:v.MEDIA_LOOP_REQUEST,detail:n!=null}):e===U.NO_VOLUME_PREF&&n!==t?(d=V(this,H))==null||d.dispatch({type:`optionschangerequest`,detail:{noVolumePref:this.hasAttribute(U.NO_VOLUME_PREF)}}):e===U.NO_MUTED_PREF&&n!==t&&((f=V(this,H))==null||f.dispatch({type:`optionschangerequest`,detail:{noMutedPref:this.hasAttribute(U.NO_MUTED_PREF)}}))}connectedCallback(){var e,t;this.associateElement(this),!V(this,H)&&!this.hasAttribute(U.NO_DEFAULT_STORE)&&_r(this,Tr,Er).call(this),(e=V(this,H))==null||e.dispatch({type:`documentelementchangerequest`,detail:O}),(t=V(this,H))==null||t.dispatch({type:`fullscreenelementchangerequest`,detail:this.fullscreenElement}),super.connectedCallback(),V(this,H)&&!V(this,Sr)&&gr(this,Sr,V(this,H)?.subscribe(V(this,xr))),V(this,wr)!==void 0&&V(this,H)&&this.media&&setTimeout(()=>{var e;this.media?.textTracks?.length&&((e=V(this,H))==null||e.dispatch({type:v.MEDIA_TOGGLE_SUBTITLES_REQUEST,detail:V(this,wr)}))},0),this.hasAttribute(U.NO_HOTKEYS)?this.disableHotkeys():this.enableHotkeys()}disconnectedCallback(){var e,t,n,r,i;if((e=super.disconnectedCallback)==null||e.call(this),this.disableHotkeys(),V(this,H)){let e=V(this,H).getState();gr(this,wr,!!e.mediaSubtitlesShowing?.length),(t=V(this,H))==null||t.dispatch({type:`fullscreenelementchangerequest`,detail:void 0}),(n=V(this,H))==null||n.dispatch({type:`documentelementchangerequest`,detail:void 0}),(r=V(this,H))==null||r.dispatch({type:v.MEDIA_TOGGLE_SUBTITLES_REQUEST,detail:!1})}V(this,Sr)&&((i=V(this,Sr))==null||i.call(this),gr(this,Sr,void 0)),this.unassociateElement(this),V(this,br)&&(V(this,br).remove(),gr(this,br,null))}mediaSetCallback(e){var t;super.mediaSetCallback(e),(t=V(this,H))==null||t.dispatch({type:`mediaelementchangerequest`,detail:e}),e.hasAttribute(`tabindex`)||(e.tabIndex=-1)}mediaUnsetCallback(e){var t;super.mediaUnsetCallback(e),(t=V(this,H))==null||t.dispatch({type:`mediaelementchangerequest`,detail:void 0})}propagateMediaState(e,t){Yr(this.mediaStateReceivers,e,t)}associateElement(e){if(!e)return;let{associatedElementSubscriptions:t}=this;if(t.has(e))return;let n=Xr(e,this.registerMediaStateReceiver.bind(this),this.unregisterMediaStateReceiver.bind(this));Object.values(v).forEach(t=>{e.addEventListener(t,V(this,Cr))}),t.set(e,n)}unassociateElement(e){if(!e)return;let{associatedElementSubscriptions:t}=this;t.has(e)&&(t.get(e)(),t.delete(e),Object.values(v).forEach(t=>{e.removeEventListener(t,V(this,Cr))}))}registerMediaStateReceiver(e){if(!e)return;let t=this.mediaStateReceivers;t.indexOf(e)>-1||(t.push(e),V(this,H)&&Object.entries(V(this,H).getState()).forEach(([t,n])=>{Yr([e],t,n)}))}unregisterMediaStateReceiver(e){let t=this.mediaStateReceivers,n=t.indexOf(e);n<0||t.splice(n,1)}enableHotkeys(){this.addEventListener(`keydown`,_r(this,Or,kr))}disableHotkeys(){this.removeEventListener(`keydown`,_r(this,Or,kr)),this.removeEventListener(`keyup`,V(this,Dr))}get hotkeys(){return V(this,vr)}set hotkeys(e){F(this,U.HOTKEYS,e)}keyboardShortcutHandler(e){let t=e.target;if((t.getAttribute(U.KEYS_USED)?.split(` `)??t?.keysUsed??[]).map(e=>e===`Space`?` `:e).filter(Boolean).includes(e.key))return;let n,r,i;if(!V(this,vr).contains(`no${e.key.toLowerCase()}`)&&!(e.key===` `&&V(this,vr).contains(`nospace`))&&(!e.shiftKey||e.key!==`/`&&e.key!==`?`||!V(this,vr).contains(`noshift+/`)))switch(e.key){case` `:case`k`:n=V(this,H).getState().mediaPaused?v.MEDIA_PLAY_REQUEST:v.MEDIA_PAUSE_REQUEST,this.dispatchEvent(new D.CustomEvent(n,{composed:!0,bubbles:!0}));break;case`m`:n=this.mediaStore.getState().mediaVolumeLevel===`off`?v.MEDIA_UNMUTE_REQUEST:v.MEDIA_MUTE_REQUEST,this.dispatchEvent(new D.CustomEvent(n,{composed:!0,bubbles:!0}));break;case`f`:n=this.mediaStore.getState().mediaIsFullscreen?v.MEDIA_EXIT_FULLSCREEN_REQUEST:v.MEDIA_ENTER_FULLSCREEN_REQUEST,this.dispatchEvent(new D.CustomEvent(n,{composed:!0,bubbles:!0}));break;case`c`:this.dispatchEvent(new D.CustomEvent(v.MEDIA_TOGGLE_SUBTITLES_REQUEST,{composed:!0,bubbles:!0}));break;case`ArrowLeft`:case`j`:{let e=this.hasAttribute(U.KEYBOARD_BACKWARD_SEEK_OFFSET)?+this.getAttribute(U.KEYBOARD_BACKWARD_SEEK_OFFSET):Nr;r=Math.max((this.mediaStore.getState().mediaCurrentTime??0)-e,0),i=new D.CustomEvent(v.MEDIA_SEEK_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`ArrowRight`:case`l`:{let e=this.hasAttribute(U.KEYBOARD_FORWARD_SEEK_OFFSET)?+this.getAttribute(U.KEYBOARD_FORWARD_SEEK_OFFSET):Nr;r=Math.max((this.mediaStore.getState().mediaCurrentTime??0)+e,0),i=new D.CustomEvent(v.MEDIA_SEEK_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`ArrowUp`:{let e=this.hasAttribute(U.KEYBOARD_UP_VOLUME_STEP)?+this.getAttribute(U.KEYBOARD_UP_VOLUME_STEP):Pr;r=Math.min((this.mediaStore.getState().mediaVolume??1)+e,1),i=new D.CustomEvent(v.MEDIA_VOLUME_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`ArrowDown`:{let e=this.hasAttribute(U.KEYBOARD_DOWN_VOLUME_STEP)?+this.getAttribute(U.KEYBOARD_DOWN_VOLUME_STEP):Pr;r=Math.max((this.mediaStore.getState().mediaVolume??1)-e,0),i=new D.CustomEvent(v.MEDIA_VOLUME_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`<`:{let e=this.mediaStore.getState().mediaPlaybackRate??1;r=Math.max(e-Fr,Ir).toFixed(2),i=new D.CustomEvent(v.MEDIA_PLAYBACK_RATE_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`>`:{let e=this.mediaStore.getState().mediaPlaybackRate??1;r=Math.min(e+Fr,Lr).toFixed(2),i=new D.CustomEvent(v.MEDIA_PLAYBACK_RATE_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`/`:case`?`:e.shiftKey&&_r(this,Ar,jr).call(this);break;case`p`:n=this.mediaStore.getState().mediaIsPip?v.MEDIA_EXIT_PIP_REQUEST:v.MEDIA_ENTER_PIP_REQUEST,i=new D.CustomEvent(n,{composed:!0,bubbles:!0}),this.dispatchEvent(i)}}};vr=new WeakMap,yr=new WeakMap,H=new WeakMap,br=new WeakMap,xr=new WeakMap,Sr=new WeakMap,Cr=new WeakMap,wr=new WeakMap,Tr=new WeakSet,Er=function(){this.mediaStore=pr({media:this.media,fullscreenElement:this.fullscreenElement,options:{defaultSubtitles:this.hasAttribute(U.DEFAULT_SUBTITLES),defaultDuration:this.hasAttribute(U.DEFAULT_DURATION)?+this.getAttribute(U.DEFAULT_DURATION):void 0,defaultStreamType:this.getAttribute(U.DEFAULT_STREAM_TYPE)??void 0,liveEdgeOffset:this.hasAttribute(U.LIVE_EDGE_OFFSET)?+this.getAttribute(U.LIVE_EDGE_OFFSET):void 0,seekToLiveOffset:this.hasAttribute(U.SEEK_TO_LIVE_OFFSET)?+this.getAttribute(U.SEEK_TO_LIVE_OFFSET):this.hasAttribute(U.LIVE_EDGE_OFFSET)?+this.getAttribute(U.LIVE_EDGE_OFFSET):void 0,noAutoSeekToLive:this.hasAttribute(U.NO_AUTO_SEEK_TO_LIVE),noVolumePref:this.hasAttribute(U.NO_VOLUME_PREF),noMutedPref:this.hasAttribute(U.NO_MUTED_PREF),noSubtitlesLangPref:this.hasAttribute(U.NO_SUBTITLES_LANG_PREF)}})},Dr=new WeakMap,Or=new WeakSet,kr=function(e){let{metaKey:t,altKey:n,key:r,shiftKey:i}=e,a=i&&(r===`/`||r===`?`);if(a&&V(this,br)?.open){this.removeEventListener(`keyup`,V(this,Dr));return}if(t||n||!a&&!Mr.includes(r)){this.removeEventListener(`keyup`,V(this,Dr));return}let o=e.target,s=o instanceof HTMLElement&&(o.tagName.toLowerCase()===`media-volume-range`||o.tagName.toLowerCase()===`media-time-range`);[` `,`ArrowLeft`,`ArrowRight`,`ArrowUp`,`ArrowDown`].includes(r)&&!(V(this,vr).contains(`no${r.toLowerCase()}`)||r===` `&&V(this,vr).contains(`nospace`))&&!s&&e.preventDefault(),this.addEventListener(`keyup`,V(this,Dr),{once:!0})},Ar=new WeakSet,jr=function(){V(this,br)||(gr(this,br,O.createElement(`media-keyboard-shortcuts-dialog`)),this.appendChild(V(this,br))),V(this,br).open=!0};var zr=Object.values(b),Br=Object.values(Ne),Vr=e=>{var t;let{observedAttributes:n}=e.constructor;!n&&e.nodeName?.includes(`-`)&&(D.customElements.upgrade(e),{observedAttributes:n}=e.constructor);let r=((t=(e?.getAttribute)?.call(e,y.MEDIA_CHROME_ATTRIBUTES))?.split)?.call(t,/\s+/);return Array.isArray(n||r)?(n||r).filter(e=>zr.includes(e)):[]},Hr=e=>(e.nodeName?.includes(`-`)&&D.customElements.get(e.nodeName?.toLowerCase())&&!(e instanceof D.customElements.get(e.nodeName.toLowerCase()))&&D.customElements.upgrade(e),Br.some(t=>t in e)),Ur=e=>Hr(e)||!!Vr(e).length,Wr=e=>(e?.join)?.call(e,`:`),Gr={[b.MEDIA_SUBTITLES_LIST]:On,[b.MEDIA_SUBTITLES_SHOWING]:On,[b.MEDIA_SEEKABLE]:Wr,[b.MEDIA_BUFFERED]:e=>e?.map(Wr).join(` `),[b.MEDIA_PREVIEW_COORDS]:e=>e?.join(` `),[b.MEDIA_RENDITION_LIST]:Re,[b.MEDIA_AUDIO_TRACK_LIST]:Be},Kr=async(e,t,n)=>{if(e.isConnected||await Ue(0),typeof n==`boolean`||n==null)return N(e,t,n);if(typeof n==`number`)return j(e,t,n);if(typeof n==`string`)return F(e,t,n);if(Array.isArray(n)&&!n.length)return e.removeAttribute(t);let r=Gr[t]?.call(Gr,n)??n;return e.setAttribute(t,r)},qr=e=>!!e.closest?.call(e,`*[slot="media"]`),Jr=(e,t)=>{if(qr(e))return;let n=(e,t)=>{Ur(e)&&t(e);let{children:n=[]}=e??{},r=e?.shadowRoot?.children??[];[...n,...r].forEach(e=>Jr(e,t))},r=e?.nodeName.toLowerCase();if(r.includes(`-`)&&!Ur(e)){D.customElements.whenDefined(r).then(()=>{n(e,t)});return}n(e,t)},Yr=(e,t,n)=>{e.forEach(e=>{if(t in e){e[t]=n;return}let r=Vr(e),i=t.toLowerCase();r.includes(i)&&Kr(e,i,n)})},Xr=(e,t,n)=>{Jr(e,t);let r=e=>{t(e?.composedPath()[0]??e.target)},i=e=>{n(e?.composedPath()[0]??e.target)};e.addEventListener(v.REGISTER_MEDIA_STATE_RECEIVER,r),e.addEventListener(v.UNREGISTER_MEDIA_STATE_RECEIVER,i);let a=e=>{e.forEach(e=>{let{addedNodes:r=[],removedNodes:i=[],type:a,target:o,attributeName:s}=e;a===`childList`?(Array.prototype.forEach.call(r,e=>Jr(e,t)),Array.prototype.forEach.call(i,e=>Jr(e,n))):a===`attributes`&&s===y.MEDIA_CHROME_ATTRIBUTES&&(Ur(o)?t(o):n(o))})},o=[],s=e=>{let r=e.target;r.name!==`media`&&(o.forEach(e=>Jr(e,n)),o=[...r.assignedElements({flatten:!0})],o.forEach(e=>Jr(e,t)))};e.addEventListener(`slotchange`,s);let c=new MutationObserver(a);return c.observe(e,{childList:!0,attributes:!0,subtree:!0}),()=>{Jr(e,n),e.removeEventListener(`slotchange`,s),c.disconnect(),e.removeEventListener(v.REGISTER_MEDIA_STATE_RECEIVER,r),e.removeEventListener(v.UNREGISTER_MEDIA_STATE_RECEIVER,i)}};D.customElements.get(`media-controller`)||D.customElements.define(`media-controller`,Rr);var Zr=Rr,Qr={PLACEMENT:`placement`,BOUNDS:`bounds`};function $r(e){return`
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
  `}var ei=class extends D.HTMLElement{constructor(){if(super(),this.updateXOffset=()=>{if(!St(this,{checkOpacity:!1,checkVisibilityCSS:!1}))return;let e=this.placement;if(e===`left`||e===`right`){this.style.removeProperty(`--media-tooltip-offset-x`);return}let t=getComputedStyle(this),n=yt(this,`#`+this.bounds)??pt(this);if(!n)return;let{x:r,width:i}=n.getBoundingClientRect(),{x:a,width:o}=this.getBoundingClientRect(),s=a+o,c=r+i,l=t.getPropertyValue(`--media-tooltip-offset-x`),u=l?parseFloat(l.replace(`px`,``)):0,d=t.getPropertyValue(`--media-tooltip-container-margin`),f=d?parseFloat(d.replace(`px`,``)):0,p=a-r+u-f,m=s-c+u+f;if(p<0){this.style.setProperty(`--media-tooltip-offset-x`,`${p}px`);return}if(m>0){this.style.setProperty(`--media-tooltip-offset-x`,`${m}px`);return}this.style.removeProperty(`--media-tooltip-offset-x`)},!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=ft(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}if(this.arrowEl=this.shadowRoot.querySelector(`#arrow`),Object.prototype.hasOwnProperty.call(this,`placement`)){let e=this.placement;delete this.placement,this.placement=e}}static get observedAttributes(){return[Qr.PLACEMENT,Qr.BOUNDS]}get placement(){return P(this,Qr.PLACEMENT)}set placement(e){F(this,Qr.PLACEMENT,e)}get bounds(){return P(this,Qr.BOUNDS)}set bounds(e){F(this,Qr.BOUNDS,e)}};ei.shadowRootOptions={mode:`open`},ei.getTemplateHTML=$r,D.customElements.get(`media-tooltip`)||D.customElements.define(`media-tooltip`,ei);var ti=ei,ni=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},W=(e,t,n)=>(ni(e,t,`read from private field`),n?n.call(e):t.get(e)),ri=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},ii=(e,t,n,r)=>(ni(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),ai=(e,t,n)=>(ni(e,t,`access private method`),n),oi,si,ci,li,ui,di,fi,pi={TOOLTIP_PLACEMENT:`tooltipplacement`,DISABLED:`disabled`,NO_TOOLTIP:`notooltip`};function mi(e,t={}){return`
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
        <template shadowrootmode="${ti.shadowRootOptions.mode}">
          ${ti.getTemplateHTML({})}
        </template>
        <slot name="tooltip-content">
          ${this.getTooltipContentHTML(e)}
        </slot>
      </media-tooltip>
    </slot>
  `}function hi(e,t){return`
    <slot></slot>
  `}function gi(){return``}var G=class extends D.HTMLElement{constructor(){if(super(),ri(this,di),ri(this,oi,void 0),this.preventClick=!1,this.tooltipEl=null,ri(this,si,e=>{this.preventClick||this.handleClick(e),setTimeout(W(this,ci),0)}),ri(this,ci,()=>{var e,t;(t=(e=this.tooltipEl)?.updateXOffset)==null||t.call(e)}),ri(this,li,e=>{let{key:t}=e;if(!this.keysUsed.includes(t)){this.removeEventListener(`keyup`,W(this,li));return}this.preventClick||this.handleClick(e)}),ri(this,ui,e=>{let{metaKey:t,altKey:n,key:r}=e;if(t||n||!this.keysUsed.includes(r)){this.removeEventListener(`keyup`,W(this,li));return}this.addEventListener(`keyup`,W(this,li),{once:!0})}),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=ft(this.attributes),t=this.constructor.getTemplateHTML(e);this.shadowRoot.setHTMLUnsafe?this.shadowRoot.setHTMLUnsafe(t):this.shadowRoot.innerHTML=t}this.tooltipEl=this.shadowRoot.querySelector(`media-tooltip`)}static get observedAttributes(){return[`disabled`,pi.TOOLTIP_PLACEMENT,y.MEDIA_CONTROLLER,b.MEDIA_LANG]}enable(){this.addEventListener(`click`,W(this,si)),this.addEventListener(`keydown`,W(this,ui)),this.tabIndex=0}disable(){this.removeEventListener(`click`,W(this,si)),this.removeEventListener(`keydown`,W(this,ui)),this.removeEventListener(`keyup`,W(this,li)),this.tabIndex=-1}attributeChangedCallback(e,t,n){var r,i,a,o;e===y.MEDIA_CONTROLLER?(t&&((i=(r=W(this,oi))?.unassociateElement)==null||i.call(r,this),ii(this,oi,null)),n&&this.isConnected&&(ii(this,oi,this.getRootNode()?.getElementById(n)),(o=(a=W(this,oi))?.associateElement)==null||o.call(a,this))):e===`disabled`&&n!==t?n==null?this.enable():this.disable():e===pi.TOOLTIP_PLACEMENT&&this.tooltipEl&&n!==t?this.tooltipEl.placement=n:e===b.MEDIA_LANG&&(this.shadowRoot.querySelector(`slot[name="tooltip-content"]`).innerHTML=this.constructor.getTooltipContentHTML()),W(this,ci).call(this)}connectedCallback(){var e,t;let{style:n}=k(this.shadowRoot,`:host`);n.setProperty(`display`,`var(--media-control-display, var(--${this.localName}-display, inline-flex))`),this.hasAttribute(`disabled`)?this.disable():this.enable(),this.setAttribute(`role`,`button`);let r=this.getAttribute(y.MEDIA_CONTROLLER);r&&(ii(this,oi,this.getRootNode()?.getElementById(r)),(t=(e=W(this,oi))?.associateElement)==null||t.call(e,this)),D.customElements.whenDefined(`media-tooltip`).then(()=>ai(this,di,fi).call(this))}disconnectedCallback(){var e,t;this.disable(),(t=(e=W(this,oi))?.unassociateElement)==null||t.call(e,this),ii(this,oi,null),this.removeEventListener(`mouseenter`,W(this,ci)),this.removeEventListener(`focus`,W(this,ci)),this.removeEventListener(`click`,W(this,si))}get keysUsed(){return[`Enter`,` `]}get tooltipPlacement(){return P(this,pi.TOOLTIP_PLACEMENT)}set tooltipPlacement(e){F(this,pi.TOOLTIP_PLACEMENT,e)}get mediaController(){return P(this,y.MEDIA_CONTROLLER)}set mediaController(e){F(this,y.MEDIA_CONTROLLER,e)}get disabled(){return M(this,pi.DISABLED)}set disabled(e){N(this,pi.DISABLED,e)}get noTooltip(){return M(this,pi.NO_TOOLTIP)}set noTooltip(e){N(this,pi.NO_TOOLTIP,e)}handleClick(e){}};oi=new WeakMap,si=new WeakMap,ci=new WeakMap,li=new WeakMap,ui=new WeakMap,di=new WeakSet,fi=function(){this.addEventListener(`mouseenter`,W(this,ci)),this.addEventListener(`focus`,W(this,ci)),this.addEventListener(`click`,W(this,si));let e=this.tooltipPlacement;e&&this.tooltipEl&&(this.tooltipEl.placement=e)},G.shadowRootOptions={mode:`open`},G.getTemplateHTML=mi,G.getSlotTemplateHTML=hi,G.getTooltipContentHTML=gi,D.customElements.get(`media-chrome-button`)||D.customElements.define(`media-chrome-button`,G);var _i=G,vi=`<svg aria-hidden="true" viewBox="0 0 26 24">
  <path d="M22.13 3H3.87a.87.87 0 0 0-.87.87v13.26a.87.87 0 0 0 .87.87h3.4L9 16H5V5h16v11h-4l1.72 2h3.4a.87.87 0 0 0 .87-.87V3.87a.87.87 0 0 0-.86-.87Zm-8.75 11.44a.5.5 0 0 0-.76 0l-4.91 5.73a.5.5 0 0 0 .38.83h9.82a.501.501 0 0 0 .38-.83l-4.91-5.73Z"/>
</svg>
`;function yi(e){return`
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
      <slot name="enter">${vi}</slot>
      <slot name="exit">${vi}</slot>
    </slot>
  `}function bi(){return`
    <slot name="tooltip-enter">${E(`start airplay`)}</slot>
    <slot name="tooltip-exit">${E(`stop airplay`)}</slot>
  `}var xi=e=>{let t=e.mediaIsAirplaying?E(`stop airplay`):E(`start airplay`);e.setAttribute(`aria-label`,t)},Si=class extends G{static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_IS_AIRPLAYING,b.MEDIA_AIRPLAY_UNAVAILABLE]}connectedCallback(){super.connectedCallback(),xi(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===b.MEDIA_IS_AIRPLAYING&&xi(this)}get mediaIsAirplaying(){return M(this,b.MEDIA_IS_AIRPLAYING)}set mediaIsAirplaying(e){N(this,b.MEDIA_IS_AIRPLAYING,e)}get mediaAirplayUnavailable(){return P(this,b.MEDIA_AIRPLAY_UNAVAILABLE)}set mediaAirplayUnavailable(e){F(this,b.MEDIA_AIRPLAY_UNAVAILABLE,e)}handleClick(){let e=new D.CustomEvent(v.MEDIA_AIRPLAY_REQUEST,{composed:!0,bubbles:!0});this.dispatchEvent(e)}};Si.getSlotTemplateHTML=yi,Si.getTooltipContentHTML=bi,D.customElements.get(`media-airplay-button`)||D.customElements.define(`media-airplay-button`,Si);var Ci=Si,wi=`<svg aria-hidden="true" viewBox="0 0 26 24">
  <path d="M22.83 5.68a2.58 2.58 0 0 0-2.3-2.5c-3.62-.24-11.44-.24-15.06 0a2.58 2.58 0 0 0-2.3 2.5c-.23 4.21-.23 8.43 0 12.64a2.58 2.58 0 0 0 2.3 2.5c3.62.24 11.44.24 15.06 0a2.58 2.58 0 0 0 2.3-2.5c.23-4.21.23-8.43 0-12.64Zm-11.39 9.45a3.07 3.07 0 0 1-1.91.57 3.06 3.06 0 0 1-2.34-1 3.75 3.75 0 0 1-.92-2.67 3.92 3.92 0 0 1 .92-2.77 3.18 3.18 0 0 1 2.43-1 2.94 2.94 0 0 1 2.13.78c.364.359.62.813.74 1.31l-1.43.35a1.49 1.49 0 0 0-1.51-1.17 1.61 1.61 0 0 0-1.29.58 2.79 2.79 0 0 0-.5 1.89 3 3 0 0 0 .49 1.93 1.61 1.61 0 0 0 1.27.58 1.48 1.48 0 0 0 1-.37 2.1 2.1 0 0 0 .59-1.14l1.4.44a3.23 3.23 0 0 1-1.07 1.69Zm7.22 0a3.07 3.07 0 0 1-1.91.57 3.06 3.06 0 0 1-2.34-1 3.75 3.75 0 0 1-.92-2.67 3.88 3.88 0 0 1 .93-2.77 3.14 3.14 0 0 1 2.42-1 3 3 0 0 1 2.16.82 2.8 2.8 0 0 1 .73 1.31l-1.43.35a1.49 1.49 0 0 0-1.51-1.21 1.61 1.61 0 0 0-1.29.58A2.79 2.79 0 0 0 15 12a3 3 0 0 0 .49 1.93 1.61 1.61 0 0 0 1.27.58 1.44 1.44 0 0 0 1-.37 2.1 2.1 0 0 0 .6-1.15l1.4.44a3.17 3.17 0 0 1-1.1 1.7Z"/>
</svg>`,Ti=`<svg aria-hidden="true" viewBox="0 0 26 24">
  <path d="M17.73 14.09a1.4 1.4 0 0 1-1 .37 1.579 1.579 0 0 1-1.27-.58A3 3 0 0 1 15 12a2.8 2.8 0 0 1 .5-1.85 1.63 1.63 0 0 1 1.29-.57 1.47 1.47 0 0 1 1.51 1.2l1.43-.34A2.89 2.89 0 0 0 19 9.07a3 3 0 0 0-2.14-.78 3.14 3.14 0 0 0-2.42 1 3.91 3.91 0 0 0-.93 2.78 3.74 3.74 0 0 0 .92 2.66 3.07 3.07 0 0 0 2.34 1 3.07 3.07 0 0 0 1.91-.57 3.17 3.17 0 0 0 1.07-1.74l-1.4-.45c-.083.43-.3.822-.62 1.12Zm-7.22 0a1.43 1.43 0 0 1-1 .37 1.58 1.58 0 0 1-1.27-.58A3 3 0 0 1 7.76 12a2.8 2.8 0 0 1 .5-1.85 1.63 1.63 0 0 1 1.29-.57 1.47 1.47 0 0 1 1.51 1.2l1.43-.34a2.81 2.81 0 0 0-.74-1.32 2.94 2.94 0 0 0-2.13-.78 3.18 3.18 0 0 0-2.43 1 4 4 0 0 0-.92 2.78 3.74 3.74 0 0 0 .92 2.66 3.07 3.07 0 0 0 2.34 1 3.07 3.07 0 0 0 1.91-.57 3.23 3.23 0 0 0 1.07-1.74l-1.4-.45a2.06 2.06 0 0 1-.6 1.07Zm12.32-8.41a2.59 2.59 0 0 0-2.3-2.51C18.72 3.05 15.86 3 13 3c-2.86 0-5.72.05-7.53.17a2.59 2.59 0 0 0-2.3 2.51c-.23 4.207-.23 8.423 0 12.63a2.57 2.57 0 0 0 2.3 2.5c1.81.13 4.67.19 7.53.19 2.86 0 5.72-.06 7.53-.19a2.57 2.57 0 0 0 2.3-2.5c.23-4.207.23-8.423 0-12.63Zm-1.49 12.53a1.11 1.11 0 0 1-.91 1.11c-1.67.11-4.45.18-7.43.18-2.98 0-5.76-.07-7.43-.18a1.11 1.11 0 0 1-.91-1.11c-.21-4.14-.21-8.29 0-12.43a1.11 1.11 0 0 1 .91-1.11C7.24 4.56 10 4.49 13 4.49s5.76.07 7.43.18a1.11 1.11 0 0 1 .91 1.11c.21 4.14.21 8.29 0 12.43Z"/>
</svg>`;function Ei(e){return`
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
      <slot name="on">${wi}</slot>
      <slot name="off">${Ti}</slot>
    </slot>
  `}function Di(){return`
    <slot name="tooltip-enable">${E(`Enable captions`)}</slot>
    <slot name="tooltip-disable">${E(`Disable captions`)}</slot>
  `}var Oi=e=>{e.setAttribute(`aria-checked`,Nn(e).toString())},ki=class extends G{static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_SUBTITLES_LIST,b.MEDIA_SUBTITLES_SHOWING]}connectedCallback(){super.connectedCallback(),this.setAttribute(`role`,`button`),this.setAttribute(`aria-label`,E(`closed captions`)),Oi(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===b.MEDIA_SUBTITLES_SHOWING&&Oi(this)}get mediaSubtitlesList(){return Ai(this,b.MEDIA_SUBTITLES_LIST)}set mediaSubtitlesList(e){ji(this,b.MEDIA_SUBTITLES_LIST,e)}get mediaSubtitlesShowing(){return Ai(this,b.MEDIA_SUBTITLES_SHOWING)}set mediaSubtitlesShowing(e){ji(this,b.MEDIA_SUBTITLES_SHOWING,e)}handleClick(){this.dispatchEvent(new D.CustomEvent(v.MEDIA_TOGGLE_SUBTITLES_REQUEST,{composed:!0,bubbles:!0}))}};ki.getSlotTemplateHTML=Ei,ki.getTooltipContentHTML=Di;var Ai=(e,t)=>{let n=e.getAttribute(t);return n?Tn(n):[]},ji=(e,t,n)=>{if(!n?.length){e.removeAttribute(t);return}let r=On(n);e.getAttribute(t)!==r&&e.setAttribute(t,r)};D.customElements.get(`media-captions-button`)||D.customElements.define(`media-captions-button`,ki);var Mi=ki,Ni=`<svg aria-hidden="true" viewBox="0 0 24 24"><g><path class="cast_caf_icon_arch0" d="M1,18 L1,21 L4,21 C4,19.3 2.66,18 1,18 L1,18 Z"/><path class="cast_caf_icon_arch1" d="M1,14 L1,16 C3.76,16 6,18.2 6,21 L8,21 C8,17.13 4.87,14 1,14 L1,14 Z"/><path class="cast_caf_icon_arch2" d="M1,10 L1,12 C5.97,12 10,16.0 10,21 L12,21 C12,14.92 7.07,10 1,10 L1,10 Z"/><path class="cast_caf_icon_box" d="M21,3 L3,3 C1.9,3 1,3.9 1,5 L1,8 L3,8 L3,5 L21,5 L21,19 L14,19 L14,21 L21,21 C22.1,21 23,20.1 23,19 L23,5 C23,3.9 22.1,3 21,3 L21,3 Z"/></g></svg>`,Pi=`<svg aria-hidden="true" viewBox="0 0 24 24"><g><path class="cast_caf_icon_arch0" d="M1,18 L1,21 L4,21 C4,19.3 2.66,18 1,18 L1,18 Z"/><path class="cast_caf_icon_arch1" d="M1,14 L1,16 C3.76,16 6,18.2 6,21 L8,21 C8,17.13 4.87,14 1,14 L1,14 Z"/><path class="cast_caf_icon_arch2" d="M1,10 L1,12 C5.97,12 10,16.0 10,21 L12,21 C12,14.92 7.07,10 1,10 L1,10 Z"/><path class="cast_caf_icon_box" d="M21,3 L3,3 C1.9,3 1,3.9 1,5 L1,8 L3,8 L3,5 L21,5 L21,19 L14,19 L14,21 L21,21 C22.1,21 23,20.1 23,19 L23,5 C23,3.9 22.1,3 21,3 L21,3 Z"/><path class="cast_caf_icon_boxfill" d="M5,7 L5,8.63 C8,8.6 13.37,14 13.37,17 L19,17 L19,7 Z"/></g></svg>`;function Fi(e){return`
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
      <slot name="enter">${Ni}</slot>
      <slot name="exit">${Pi}</slot>
    </slot>
  `}function Ii(){return`
    <slot name="tooltip-enter">${E(`Start casting`)}</slot>
    <slot name="tooltip-exit">${E(`Stop casting`)}</slot>
  `}var Li=e=>{let t=e.mediaIsCasting?E(`stop casting`):E(`start casting`);e.setAttribute(`aria-label`,t)},Ri=class extends G{static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_IS_CASTING,b.MEDIA_CAST_UNAVAILABLE]}connectedCallback(){super.connectedCallback(),Li(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===b.MEDIA_IS_CASTING&&Li(this)}get mediaIsCasting(){return M(this,b.MEDIA_IS_CASTING)}set mediaIsCasting(e){N(this,b.MEDIA_IS_CASTING,e)}get mediaCastUnavailable(){return P(this,b.MEDIA_CAST_UNAVAILABLE)}set mediaCastUnavailable(e){F(this,b.MEDIA_CAST_UNAVAILABLE,e)}handleClick(){let e=this.mediaIsCasting?v.MEDIA_EXIT_CAST_REQUEST:v.MEDIA_ENTER_CAST_REQUEST;this.dispatchEvent(new D.CustomEvent(e,{composed:!0,bubbles:!0}))}};Ri.getSlotTemplateHTML=Fi,Ri.getTooltipContentHTML=Ii,D.customElements.get(`media-cast-button`)||D.customElements.define(`media-cast-button`,Ri);var zi=Ri,Bi=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Vi=(e,t,n)=>(Bi(e,t,`read from private field`),n?n.call(e):t.get(e)),Hi=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Ui=(e,t,n,r)=>(Bi(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),Wi=(e,t,n)=>(Bi(e,t,`access private method`),n),Gi,Ki,qi,Ji,Yi,Xi,Zi,Qi,$i,ea,ta,na,ra,ia,aa;function oa(e){return`
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
  `}function sa(e){return`
    <slot id="content"></slot>
  `}var ca={OPEN:`open`,ANCHOR:`anchor`},la=class extends D.HTMLElement{constructor(){super(),Hi(this,Ji),Hi(this,Xi),Hi(this,Qi),Hi(this,ea),Hi(this,na),Hi(this,ia),Hi(this,Gi,!1),Hi(this,Ki,null),Hi(this,qi,null)}static get observedAttributes(){return[ca.OPEN,ca.ANCHOR]}get open(){return M(this,ca.OPEN)}set open(e){N(this,ca.OPEN,e)}handleEvent(e){switch(e.type){case`invoke`:Wi(this,ea,ta).call(this,e);break;case`focusout`:Wi(this,na,ra).call(this,e);break;case`keydown`:Wi(this,ia,aa).call(this,e)}}connectedCallback(){Wi(this,Ji,Yi).call(this),this.role||=`dialog`,this.addEventListener(`invoke`,this),this.addEventListener(`focusout`,this),this.addEventListener(`keydown`,this)}disconnectedCallback(){this.removeEventListener(`invoke`,this),this.removeEventListener(`focusout`,this),this.removeEventListener(`keydown`,this)}attributeChangedCallback(e,t,n){Wi(this,Ji,Yi).call(this),e===ca.OPEN&&n!==t&&(this.open?Wi(this,Xi,Zi).call(this):Wi(this,Qi,$i).call(this))}focus(){Ui(this,Ki,bt());let e=!this.dispatchEvent(new Event(`focus`,{composed:!0,cancelable:!0})),t=!this.dispatchEvent(new Event(`focusin`,{composed:!0,bubbles:!0,cancelable:!0}));e||t||this.querySelector(`[autofocus], [tabindex]:not([tabindex="-1"]), [role="menu"]`)?.focus()}get keysUsed(){return[`Escape`,`Tab`]}};Gi=new WeakMap,Ki=new WeakMap,qi=new WeakMap,Ji=new WeakSet,Yi=function(){if(!Vi(this,Gi)&&(Ui(this,Gi,!0),!this.shadowRoot)){this.attachShadow(this.constructor.shadowRootOptions);let e=ft(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e),queueMicrotask(()=>{let{style:e}=k(this.shadowRoot,`:host`);e.setProperty(`transition`,`display .15s, visibility .15s, opacity .15s ease-in, transform .15s ease-in`)})}},Xi=new WeakSet,Zi=function(){var e;(e=Vi(this,qi))==null||e.setAttribute(`aria-expanded`,`true`),this.dispatchEvent(new Event(`open`,{composed:!0,bubbles:!0})),this.addEventListener(`transitionend`,()=>this.focus(),{once:!0})},Qi=new WeakSet,$i=function(){var e;(e=Vi(this,qi))==null||e.setAttribute(`aria-expanded`,`false`),this.dispatchEvent(new Event(`close`,{composed:!0,bubbles:!0}))},ea=new WeakSet,ta=function(e){Ui(this,qi,e.relatedTarget),vt(this,e.relatedTarget)||(this.open=!this.open)},na=new WeakSet,ra=function(e){var t;vt(this,e.relatedTarget)||((t=Vi(this,Ki))==null||t.focus(),Vi(this,qi)&&Vi(this,qi)!==e.relatedTarget&&this.open&&(this.open=!1))},ia=new WeakSet,aa=function(e){var t,n,r,i,a;let{key:o,ctrlKey:s,altKey:c,metaKey:l}=e;s||c||l||this.keysUsed.includes(o)&&(e.preventDefault(),e.stopPropagation(),o===`Tab`?(e.shiftKey?(n=(t=this.previousElementSibling)?.focus)==null||n.call(t):(i=(r=this.nextElementSibling)?.focus)==null||i.call(r),this.blur()):o===`Escape`&&((a=Vi(this,Ki))==null||a.focus(),this.open=!1))},la.shadowRootOptions={mode:`open`},la.getTemplateHTML=oa,la.getSlotTemplateHTML=sa,D.customElements.get(`media-chrome-dialog`)||D.customElements.define(`media-chrome-dialog`,la);var ua=la,da=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},K=(e,t,n)=>(da(e,t,`read from private field`),n?n.call(e):t.get(e)),q=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},fa=(e,t,n,r)=>(da(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),pa=(e,t,n)=>(da(e,t,`access private method`),n),ma,ha,ga,_a,va,ya,ba,xa,Sa,Ca,wa,Ta,Ea,Da,Oa,ka,Aa,ja,Ma,Na,Pa,Fa,Ia,La,Ra;function za(e){return`
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
  `}function Ba(e){return``}var Va=class extends D.HTMLElement{constructor(){if(super(),q(this,Ca),q(this,Ta),q(this,Da),q(this,ka),q(this,ja),q(this,Na),q(this,Fa),q(this,La),q(this,ma,void 0),q(this,ha,void 0),q(this,ga,void 0),q(this,_a,void 0),q(this,va,{}),q(this,ya,[]),q(this,ba,()=>{if(this.range.matches(`:focus-visible`)){let{style:e}=k(this.shadowRoot,`:host`);e.setProperty(`--_focus-visible-box-shadow`,`var(--_focus-box-shadow)`)}}),q(this,xa,()=>{let{style:e}=k(this.shadowRoot,`:host`);e.removeProperty(`--_focus-visible-box-shadow`)}),q(this,Sa,()=>{let e=this.shadowRoot.querySelector(`#segments-clipping`);e&&e.parentNode.append(e)}),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=ft(this.attributes),t=this.constructor.getTemplateHTML(e);this.shadowRoot.setHTMLUnsafe?this.shadowRoot.setHTMLUnsafe(t):this.shadowRoot.innerHTML=t}this.container=this.shadowRoot.querySelector(`#container`),fa(this,ga,this.shadowRoot.querySelector(`#startpoint`)),fa(this,_a,this.shadowRoot.querySelector(`#endpoint`)),this.range=this.shadowRoot.querySelector(`#range`),this.appearance=this.shadowRoot.querySelector(`#appearance`)}static get observedAttributes(){return[`disabled`,`aria-disabled`,y.MEDIA_CONTROLLER]}attributeChangedCallback(e,t,n){var r,i,a,o;e===y.MEDIA_CONTROLLER?(t&&((i=(r=K(this,ma))?.unassociateElement)==null||i.call(r,this),fa(this,ma,null)),n&&this.isConnected&&(fa(this,ma,this.getRootNode()?.getElementById(n)),(o=(a=K(this,ma))?.associateElement)==null||o.call(a,this))):(e===`disabled`||e===`aria-disabled`&&t!==n)&&(n==null?(this.range.removeAttribute(e),pa(this,Ta,Ea).call(this)):(this.range.setAttribute(e,n),pa(this,Da,Oa).call(this)))}connectedCallback(){var e,t;let{style:n}=k(this.shadowRoot,`:host`);n.setProperty(`display`,`var(--media-control-display, var(--${this.localName}-display, inline-flex))`),K(this,va).pointer=k(this.shadowRoot,`#pointer`),K(this,va).progress=k(this.shadowRoot,`#progress`),K(this,va).thumb=k(this.shadowRoot,`#thumb, ::slotted([slot="thumb"])`),K(this,va).activeSegment=k(this.shadowRoot,`#segments-clipping rect:nth-child(0)`);let r=this.getAttribute(y.MEDIA_CONTROLLER);r&&(fa(this,ma,this.getRootNode()?.getElementById(r)),(t=(e=K(this,ma))?.associateElement)==null||t.call(e,this)),this.updateBar(),this.shadowRoot.addEventListener(`focusin`,K(this,ba)),this.shadowRoot.addEventListener(`focusout`,K(this,xa)),pa(this,Ta,Ea).call(this),ut(this.container,K(this,Sa))}disconnectedCallback(){var e,t;pa(this,Da,Oa).call(this),(t=(e=K(this,ma))?.unassociateElement)==null||t.call(e,this),fa(this,ma,null),this.shadowRoot.removeEventListener(`focusin`,K(this,ba)),this.shadowRoot.removeEventListener(`focusout`,K(this,xa)),dt(this.container,K(this,Sa))}updatePointerBar(e){var t;(t=K(this,va).pointer)==null||t.style.setProperty(`width`,`${this.getPointerRatio(e)*100}%`)}updateBar(){var e,t;let n=this.range.valueAsNumber*100;(e=K(this,va).progress)==null||e.style.setProperty(`width`,`${n}%`),(t=K(this,va).thumb)==null||t.style.setProperty(`left`,`${n}%`)}updateSegments(e){let t=this.shadowRoot.querySelector(`#segments-clipping`);if(t.textContent=``,this.container.classList.toggle(`segments`,!!e?.length),!e?.length)return;let n=[...new Set([+this.range.min,...e.flatMap(e=>[e.start,e.end]),+this.range.max])];fa(this,ya,[...n]);let r=n.pop();for(let[e,i]of n.entries()){let[a,o]=[e===0,e===n.length-1],s=a?`calc(var(--segments-gap) / -1)`:`${i*100}%`,c=`calc(${((o?r:n[e+1])-i)*100}%${a||o?``:` - var(--segments-gap)`})`,l=O.createElementNS(`http://www.w3.org/2000/svg`,`rect`),u=Tt(this.shadowRoot,`#segments-clipping rect:nth-child(${e+1})`);u.style.setProperty(`x`,s),u.style.setProperty(`width`,c),t.append(l)}}getPointerRatio(e){return Ct(e.clientX,e.clientY,K(this,ga).getBoundingClientRect(),K(this,_a).getBoundingClientRect())}get dragging(){return this.hasAttribute(`dragging`)}handleEvent(e){switch(e.type){case`pointermove`:pa(this,La,Ra).call(this,e);break;case`input`:this.updateBar();break;case`pointerenter`:pa(this,ja,Ma).call(this,e);break;case`pointerdown`:pa(this,ka,Aa).call(this,e);break;case`pointerup`:pa(this,Na,Pa).call(this);break;case`pointerleave`:pa(this,Fa,Ia).call(this)}}get keysUsed(){return[`ArrowUp`,`ArrowRight`,`ArrowDown`,`ArrowLeft`]}};ma=new WeakMap,ha=new WeakMap,ga=new WeakMap,_a=new WeakMap,va=new WeakMap,ya=new WeakMap,ba=new WeakMap,xa=new WeakMap,Sa=new WeakMap,Ca=new WeakSet,wa=function(e){let t=K(this,va).activeSegment;if(!t)return;let n=this.getPointerRatio(e),r=`#segments-clipping rect:nth-child(${K(this,ya).findIndex((e,t,r)=>{let i=r[t+1];return i!=null&&n>=e&&n<=i})+1})`;(t.selectorText!=r||!t.style.transform)&&(t.selectorText=r,t.style.setProperty(`transform`,`var(--media-range-segment-hover-transform, scaleY(2))`))},Ta=new WeakSet,Ea=function(){!this.hasAttribute(`disabled`)&&this.isConnected&&(this.addEventListener(`input`,this),this.addEventListener(`pointerdown`,this),this.addEventListener(`pointerenter`,this))},Da=new WeakSet,Oa=function(){var e,t;this.removeEventListener(`input`,this),this.removeEventListener(`pointerdown`,this),this.removeEventListener(`pointerenter`,this),this.removeEventListener(`pointerleave`,this),(e=D.window)==null||e.removeEventListener(`pointerup`,this),(t=D.window)==null||t.removeEventListener(`pointermove`,this)},ka=new WeakSet,Aa=function(e){var t;fa(this,ha,e.composedPath().includes(this.range)),(t=D.window)==null||t.addEventListener(`pointerup`,this,{once:!0})},ja=new WeakSet,Ma=function(e){var t;e.pointerType!==`mouse`&&pa(this,ka,Aa).call(this,e),this.addEventListener(`pointerleave`,this,{once:!0}),(t=D.window)==null||t.addEventListener(`pointermove`,this)},Na=new WeakSet,Pa=function(){var e;(e=D.window)==null||e.removeEventListener(`pointerup`,this),this.toggleAttribute(`dragging`,!1),this.range.disabled=this.hasAttribute(`disabled`)},Fa=new WeakSet,Ia=function(){var e,t;this.removeEventListener(`pointerleave`,this),(e=D.window)==null||e.removeEventListener(`pointermove`,this),this.toggleAttribute(`dragging`,!1),this.range.disabled=this.hasAttribute(`disabled`),(t=K(this,va).activeSegment)==null||t.style.removeProperty(`transform`)},La=new WeakSet,Ra=function(e){(e.pointerType!==`pen`||e.buttons!==0)&&(this.toggleAttribute(`dragging`,e.buttons===1||e.pointerType!==`mouse`),this.updatePointerBar(e),pa(this,Ca,wa).call(this,e),this.dragging&&(e.pointerType!==`mouse`||!K(this,ha))&&(this.range.disabled=!0,this.range.valueAsNumber=this.getPointerRatio(e),this.range.dispatchEvent(new Event(`input`,{bubbles:!0,composed:!0}))))},Va.shadowRootOptions={mode:`open`},Va.getTemplateHTML=za,Va.getContainerTemplateHTML=Ba,D.customElements.get(`media-chrome-range`)||D.customElements.define(`media-chrome-range`,Va);var Ha=Va,Ua=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Wa=(e,t,n)=>(Ua(e,t,`read from private field`),n?n.call(e):t.get(e)),Ga=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Ka=(e,t,n,r)=>(Ua(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),qa;function Ja(e){return`
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
  `}var Ya=class extends D.HTMLElement{constructor(){if(super(),Ga(this,qa,void 0),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=ft(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}}static get observedAttributes(){return[y.MEDIA_CONTROLLER]}attributeChangedCallback(e,t,n){var r,i,a,o;e===y.MEDIA_CONTROLLER&&(t&&((i=(r=Wa(this,qa))?.unassociateElement)==null||i.call(r,this),Ka(this,qa,null)),n&&this.isConnected&&(Ka(this,qa,this.getRootNode()?.getElementById(n)),(o=(a=Wa(this,qa))?.associateElement)==null||o.call(a,this)))}connectedCallback(){var e,t;let n=this.getAttribute(y.MEDIA_CONTROLLER);n&&(Ka(this,qa,this.getRootNode()?.getElementById(n)),(t=(e=Wa(this,qa))?.associateElement)==null||t.call(e,this))}disconnectedCallback(){var e,t;(t=(e=Wa(this,qa))?.unassociateElement)==null||t.call(e,this),Ka(this,qa,null)}};qa=new WeakMap,Ya.shadowRootOptions={mode:`open`},Ya.getTemplateHTML=Ja,D.customElements.get(`media-control-bar`)||D.customElements.define(`media-control-bar`,Ya);var Xa=Ya,Za=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Qa=(e,t,n)=>(Za(e,t,`read from private field`),n?n.call(e):t.get(e)),$a=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},eo=(e,t,n,r)=>(Za(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),to;function no(e,t={}){return`
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
  `}function ro(e,t){return`
    <slot></slot>
  `}var io=class extends D.HTMLElement{constructor(){if(super(),$a(this,to,void 0),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=ft(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}}static get observedAttributes(){return[y.MEDIA_CONTROLLER]}attributeChangedCallback(e,t,n){var r,i,a,o;e===y.MEDIA_CONTROLLER&&(t&&((i=(r=Qa(this,to))?.unassociateElement)==null||i.call(r,this),eo(this,to,null)),n&&this.isConnected&&(eo(this,to,this.getRootNode()?.getElementById(n)),(o=(a=Qa(this,to))?.associateElement)==null||o.call(a,this)))}connectedCallback(){var e,t;let{style:n}=k(this.shadowRoot,`:host`);n.setProperty(`display`,`var(--media-control-display, var(--${this.localName}-display, inline-flex))`);let r=this.getAttribute(y.MEDIA_CONTROLLER);r&&(eo(this,to,this.getRootNode()?.getElementById(r)),(t=(e=Qa(this,to))?.associateElement)==null||t.call(e,this))}disconnectedCallback(){var e,t;(t=(e=Qa(this,to))?.unassociateElement)==null||t.call(e,this),eo(this,to,null)}};to=new WeakMap,io.shadowRootOptions={mode:`open`},io.getTemplateHTML=no,io.getSlotTemplateHTML=ro,D.customElements.get(`media-text-display`)||D.customElements.define(`media-text-display`,io);var ao=io,oo=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},so=(e,t,n)=>(oo(e,t,`read from private field`),n?n.call(e):t.get(e)),co=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},lo=(e,t,n,r)=>(oo(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),uo;function fo(e,t){return`
    <slot>${Qe(t.mediaDuration)}</slot>
  `}var po=class extends io{constructor(){super(),co(this,uo,void 0),lo(this,uo,this.shadowRoot.querySelector(`slot`)),so(this,uo).textContent=Qe(this.mediaDuration??0)}static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_DURATION]}attributeChangedCallback(e,t,n){e===b.MEDIA_DURATION&&(so(this,uo).textContent=Qe(+n)),super.attributeChangedCallback(e,t,n)}get mediaDuration(){return A(this,b.MEDIA_DURATION)}set mediaDuration(e){j(this,b.MEDIA_DURATION,e)}};uo=new WeakMap,po.getSlotTemplateHTML=fo,D.customElements.get(`media-duration-display`)||D.customElements.define(`media-duration-display`,po);var mo=po,ho={2:E(`Network Error`),3:E(`Decode Error`),4:E(`Source Not Supported`),5:E(`Encryption Error`)},go={2:E(`A network error caused the media download to fail.`),3:E(`A media error caused playback to be aborted. The media could be corrupt or your browser does not support this format.`),4:E(`An unsupported error occurred. The server or network failed, or your browser does not support this format.`),5:E(`The media is encrypted and there are no keys to decrypt it.`)},_o=e=>e.code===1?null:{title:ho[e.code]??`Error ${e.code}`,message:go[e.code]??e.message},vo=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},yo=(e,t,n)=>(vo(e,t,`read from private field`),n?n.call(e):t.get(e)),bo=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},xo=(e,t,n,r)=>(vo(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),So;function Co(e){return`
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
    <slot name="error-${e.mediaerrorcode}" id="content">
      ${To({code:+e.mediaerrorcode,message:e.mediaerrormessage})}
    </slot>
  `}function wo(e){return e.code&&_o(e)!==null}function To(e){let{title:t,message:n}=_o(e)??{},r=``;return t&&(r+=`<slot name="error-${e.code}-title"><h3>${t}</h3></slot>`),n&&(r+=`<slot name="error-${e.code}-message"><p>${n}</p></slot>`),r}var Eo=[b.MEDIA_ERROR_CODE,b.MEDIA_ERROR_MESSAGE],Do=class extends la{constructor(){super(...arguments),bo(this,So,null)}static get observedAttributes(){return[...super.observedAttributes,...Eo]}formatErrorMessage(e){return this.constructor.formatErrorMessage(e)}attributeChangedCallback(e,t,n){if(super.attributeChangedCallback(e,t,n),!Eo.includes(e))return;let r=this.mediaError??{code:this.mediaErrorCode,message:this.mediaErrorMessage};if(this.open=wo(r),this.open&&(this.shadowRoot.querySelector(`slot`).name=`error-${this.mediaErrorCode}`,this.shadowRoot.querySelector(`#content`).innerHTML=this.formatErrorMessage(r),!this.hasAttribute(`aria-label`))){let{title:e}=_o(r);e&&this.setAttribute(`aria-label`,e)}}get mediaError(){return yo(this,So)}set mediaError(e){xo(this,So,e)}get mediaErrorCode(){return A(this,`mediaerrorcode`)}set mediaErrorCode(e){j(this,`mediaerrorcode`,e)}get mediaErrorMessage(){return P(this,`mediaerrormessage`)}set mediaErrorMessage(e){F(this,`mediaerrormessage`,e)}};So=new WeakMap,Do.getSlotTemplateHTML=Co,Do.formatErrorMessage=To,D.customElements.get(`media-error-dialog`)||D.customElements.define(`media-error-dialog`,Do);var Oo=Do,ko=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Ao=(e,t,n)=>(ko(e,t,`read from private field`),n?n.call(e):t.get(e)),jo=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Mo,No;function Po(e){return`
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
      ${Fo()}
    </slot>
  `}function Fo(){return`
    <h2>Keyboard Shortcuts</h2>
    <table class="shortcuts-table">${[{keys:[`Space`,`k`],description:`Toggle Playback`},{keys:[`m`],description:`Toggle mute`},{keys:[`f`],description:`Toggle fullscreen`},{keys:[`c`],description:`Toggle captions or subtitles, if available`},{keys:[`p`],description:`Toggle Picture in Picture`},{keys:[`←`,`j`],description:`Seek back 10s`},{keys:[`→`,`l`],description:`Seek forward 10s`},{keys:[`↑`],description:`Turn volume up`},{keys:[`↓`],description:`Turn volume down`},{keys:[`< (SHIFT+,)`],description:`Decrease playback rate`},{keys:[`> (SHIFT+.)`],description:`Increase playback rate`}].map(({keys:e,description:t})=>`
      <tr>
        <td>
          <div class="key-combo">${e.map((e,t)=>t>0?`<span class="key-separator">or</span><span class="key">${e}</span>`:`<span class="key">${e}</span>`).join(``)}</div>
        </td>
        <td class="description">${t}</td>
      </tr>
    `).join(``)}</table>
  `}var Io=class extends la{constructor(){super(...arguments),jo(this,Mo,e=>{if(!this.open)return;let t=this.shadowRoot?.querySelector(`#content`);if(!t)return;let n=e.composedPath(),r=n[0]===this||n.includes(this),i=n.includes(t);r&&!i&&(this.open=!1)}),jo(this,No,e=>{if(!this.open)return;let t=e.shiftKey&&(e.key===`/`||e.key===`?`);(e.key===`Escape`||t)&&!e.ctrlKey&&!e.altKey&&!e.metaKey&&(this.open=!1,e.preventDefault(),e.stopPropagation())})}connectedCallback(){super.connectedCallback(),this.open&&(this.addEventListener(`click`,Ao(this,Mo)),document.addEventListener(`keydown`,Ao(this,No)))}disconnectedCallback(){this.removeEventListener(`click`,Ao(this,Mo)),document.removeEventListener(`keydown`,Ao(this,No))}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===`open`&&(this.open?(this.addEventListener(`click`,Ao(this,Mo)),document.addEventListener(`keydown`,Ao(this,No))):(this.removeEventListener(`click`,Ao(this,Mo)),document.removeEventListener(`keydown`,Ao(this,No))))}};Mo=new WeakMap,No=new WeakMap,Io.getSlotTemplateHTML=Po,D.customElements.get(`media-keyboard-shortcuts-dialog`)||D.customElements.define(`media-keyboard-shortcuts-dialog`,Io);var Lo=Io,Ro=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},zo=(e,t,n)=>(Ro(e,t,`read from private field`),n?n.call(e):t.get(e)),Bo=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Vo=(e,t,n,r)=>(Ro(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),Ho,Uo=`<svg aria-hidden="true" viewBox="0 0 26 24">
  <path d="M16 3v2.5h3.5V9H22V3h-6ZM4 9h2.5V5.5H10V3H4v6Zm15.5 9.5H16V21h6v-6h-2.5v3.5ZM6.5 15H4v6h6v-2.5H6.5V15Z"/>
</svg>`,Wo=`<svg aria-hidden="true" viewBox="0 0 26 24">
  <path d="M18.5 6.5V3H16v6h6V6.5h-3.5ZM16 21h2.5v-3.5H22V15h-6v6ZM4 17.5h3.5V21H10v-6H4v2.5Zm3.5-11H4V9h6V3H7.5v3.5Z"/>
</svg>`;function Go(e){return`
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
      <slot name="enter">${Uo}</slot>
      <slot name="exit">${Wo}</slot>
    </slot>
  `}function Ko(){return`
    <slot name="tooltip-enter">${E(`Enter fullscreen mode`)}</slot>
    <slot name="tooltip-exit">${E(`Exit fullscreen mode`)}</slot>
  `}var qo=e=>{let t=e.mediaIsFullscreen?E(`exit fullscreen mode`):E(`enter fullscreen mode`);e.setAttribute(`aria-label`,t)},Jo=class extends G{constructor(){super(...arguments),Bo(this,Ho,null)}static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_IS_FULLSCREEN,b.MEDIA_FULLSCREEN_UNAVAILABLE]}connectedCallback(){super.connectedCallback(),qo(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===b.MEDIA_IS_FULLSCREEN&&qo(this)}get mediaFullscreenUnavailable(){return P(this,b.MEDIA_FULLSCREEN_UNAVAILABLE)}set mediaFullscreenUnavailable(e){F(this,b.MEDIA_FULLSCREEN_UNAVAILABLE,e)}get mediaIsFullscreen(){return M(this,b.MEDIA_IS_FULLSCREEN)}set mediaIsFullscreen(e){N(this,b.MEDIA_IS_FULLSCREEN,e)}handleClick(e){Vo(this,Ho,e);let t=zo(this,Ho)instanceof PointerEvent,n=this.mediaIsFullscreen?new D.CustomEvent(v.MEDIA_EXIT_FULLSCREEN_REQUEST,{composed:!0,bubbles:!0}):new D.CustomEvent(v.MEDIA_ENTER_FULLSCREEN_REQUEST,{composed:!0,bubbles:!0,detail:t});this.dispatchEvent(n)}};Ho=new WeakMap,Jo.getSlotTemplateHTML=Go,Jo.getTooltipContentHTML=Ko,D.customElements.get(`media-fullscreen-button`)||D.customElements.define(`media-fullscreen-button`,Jo);var Yo=Jo,{MEDIA_TIME_IS_LIVE:Xo,MEDIA_PAUSED:Zo}=b,{MEDIA_SEEK_TO_LIVE_REQUEST:Qo,MEDIA_PLAY_REQUEST:$o}=v,es=`<svg viewBox="0 0 6 12" aria-hidden="true"><circle cx="3" cy="6" r="2"></circle></svg>`;function ts(e){return`
    <style>
      :host { --media-tooltip-display: none; }
      
      slot[name=indicator] > *,
      :host ::slotted([slot=indicator]) {
        
        min-width: auto;
        fill: var(--media-live-button-icon-color, rgb(140, 140, 140));
        color: var(--media-live-button-icon-color, rgb(140, 140, 140));
      }

      :host([${Xo}]:not([${Zo}])) slot[name=indicator] > *,
      :host([${Xo}]:not([${Zo}])) ::slotted([slot=indicator]) {
        fill: var(--media-live-button-indicator-color, rgb(255, 0, 0));
        color: var(--media-live-button-indicator-color, rgb(255, 0, 0));
      }

      :host([${Xo}]:not([${Zo}])) {
        cursor: var(--media-cursor, not-allowed);
      }

      slot[name=text]{
        text-transform: uppercase;
      }

    </style>

    <slot name="indicator">${es}</slot>
    
    <slot name="spacer">&nbsp;</slot><slot name="text">${E(`live`)}</slot>
  `}var ns=e=>{let t=e.mediaPaused||!e.mediaTimeIsLive,n=E(t?`seek to live`:`playing live`);e.setAttribute(`aria-label`,n);let r=e.shadowRoot?.querySelector(`slot[name="text"]`);r&&(r.textContent=E(`live`)),t?e.removeAttribute(`aria-disabled`):e.setAttribute(`aria-disabled`,`true`)},rs=class extends G{static get observedAttributes(){return[...super.observedAttributes,Xo,Zo]}connectedCallback(){super.connectedCallback(),ns(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),ns(this)}get mediaPaused(){return M(this,b.MEDIA_PAUSED)}set mediaPaused(e){N(this,b.MEDIA_PAUSED,e)}get mediaTimeIsLive(){return M(this,b.MEDIA_TIME_IS_LIVE)}set mediaTimeIsLive(e){N(this,b.MEDIA_TIME_IS_LIVE,e)}handleClick(){(this.mediaPaused||!this.mediaTimeIsLive)&&(this.dispatchEvent(new D.CustomEvent(Qo,{composed:!0,bubbles:!0})),this.hasAttribute(Zo)&&this.dispatchEvent(new D.CustomEvent($o,{composed:!0,bubbles:!0})))}};rs.getSlotTemplateHTML=ts,D.customElements.get(`media-live-button`)||D.customElements.define(`media-live-button`,rs);var is=rs,as=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},os=(e,t,n)=>(as(e,t,`read from private field`),n?n.call(e):t.get(e)),ss=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},cs=(e,t,n,r)=>(as(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),ls,us,ds={LOADING_DELAY:`loadingdelay`,NO_AUTOHIDE:`noautohide`},fs=500,ps=`
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
`;function ms(e){return`
    <style>
      :host {
        display: var(--media-control-display, var(--media-loading-indicator-display, inline-block));
        vertical-align: middle;
        box-sizing: border-box;
        --_loading-indicator-delay: var(--media-loading-indicator-transition-delay, ${fs}ms);
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

    <slot name="icon">${ps}</slot>
    <div id="status" role="status" aria-live="polite">${E(`media loading`)}</div>
  `}var hs=class extends D.HTMLElement{constructor(){if(super(),ss(this,ls,void 0),ss(this,us,fs),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=ft(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}}static get observedAttributes(){return[y.MEDIA_CONTROLLER,b.MEDIA_PAUSED,b.MEDIA_LOADING,ds.LOADING_DELAY]}attributeChangedCallback(e,t,n){var r,i,a,o;e===ds.LOADING_DELAY&&t!==n?this.loadingDelay=Number(n):e===y.MEDIA_CONTROLLER&&(t&&((i=(r=os(this,ls))?.unassociateElement)==null||i.call(r,this),cs(this,ls,null)),n&&this.isConnected&&(cs(this,ls,this.getRootNode()?.getElementById(n)),(o=(a=os(this,ls))?.associateElement)==null||o.call(a,this)))}connectedCallback(){var e,t;let n=this.getAttribute(y.MEDIA_CONTROLLER);n&&(cs(this,ls,this.getRootNode()?.getElementById(n)),(t=(e=os(this,ls))?.associateElement)==null||t.call(e,this))}disconnectedCallback(){var e,t;(t=(e=os(this,ls))?.unassociateElement)==null||t.call(e,this),cs(this,ls,null)}get loadingDelay(){return os(this,us)}set loadingDelay(e){cs(this,us,e);let{style:t}=k(this.shadowRoot,`:host`);t.setProperty(`--_loading-indicator-delay`,`var(--media-loading-indicator-transition-delay, ${e}ms)`)}get mediaPaused(){return M(this,b.MEDIA_PAUSED)}set mediaPaused(e){N(this,b.MEDIA_PAUSED,e)}get mediaLoading(){return M(this,b.MEDIA_LOADING)}set mediaLoading(e){N(this,b.MEDIA_LOADING,e)}get mediaController(){return P(this,y.MEDIA_CONTROLLER)}set mediaController(e){F(this,y.MEDIA_CONTROLLER,e)}get noAutohide(){return M(this,ds.NO_AUTOHIDE)}set noAutohide(e){N(this,ds.NO_AUTOHIDE,e)}};ls=new WeakMap,us=new WeakMap,hs.shadowRootOptions={mode:`open`},hs.getTemplateHTML=ms,D.customElements.get(`media-loading-indicator`)||D.customElements.define(`media-loading-indicator`,hs);var gs=hs,_s=`<svg aria-hidden="true" viewBox="0 0 24 24">
  <path d="M16.5 12A4.5 4.5 0 0 0 14 8v2.18l2.45 2.45a4.22 4.22 0 0 0 .05-.63Zm2.5 0a6.84 6.84 0 0 1-.54 2.64L20 16.15A8.8 8.8 0 0 0 21 12a9 9 0 0 0-7-8.77v2.06A7 7 0 0 1 19 12ZM4.27 3 3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25A6.92 6.92 0 0 1 14 18.7v2.06A9 9 0 0 0 17.69 19l2 2.05L21 19.73l-9-9L4.27 3ZM12 4 9.91 6.09 12 8.18V4Z"/>
</svg>`,vs=`<svg aria-hidden="true" viewBox="0 0 24 24">
  <path d="M3 9v6h4l5 5V4L7 9H3Zm13.5 3A4.5 4.5 0 0 0 14 8v8a4.47 4.47 0 0 0 2.5-4Z"/>
</svg>`,ys=`<svg aria-hidden="true" viewBox="0 0 24 24">
  <path d="M3 9v6h4l5 5V4L7 9H3Zm13.5 3A4.5 4.5 0 0 0 14 8v8a4.47 4.47 0 0 0 2.5-4ZM14 3.23v2.06a7 7 0 0 1 0 13.42v2.06a9 9 0 0 0 0-17.54Z"/>
</svg>`;function bs(e){return`
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
      <slot name="off">${_s}</slot>
      <slot name="low">${vs}</slot>
      <slot name="medium">${vs}</slot>
      <slot name="high">${ys}</slot>
    </slot>
  `}function xs(){return`
    <slot name="tooltip-mute">${E(`Mute`)}</slot>
    <slot name="tooltip-unmute">${E(`Unmute`)}</slot>
  `}var Ss=e=>{let t=e.mediaVolumeLevel===`off`?E(`unmute`):E(`mute`);e.setAttribute(`aria-label`,t)},Cs=class extends G{static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_VOLUME_LEVEL]}connectedCallback(){super.connectedCallback(),Ss(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===b.MEDIA_VOLUME_LEVEL&&Ss(this)}get mediaVolumeLevel(){return P(this,b.MEDIA_VOLUME_LEVEL)}set mediaVolumeLevel(e){F(this,b.MEDIA_VOLUME_LEVEL,e)}handleClick(){let e=this.mediaVolumeLevel===`off`?v.MEDIA_UNMUTE_REQUEST:v.MEDIA_MUTE_REQUEST;this.dispatchEvent(new D.CustomEvent(e,{composed:!0,bubbles:!0}))}};Cs.getSlotTemplateHTML=bs,Cs.getTooltipContentHTML=xs,D.customElements.get(`media-mute-button`)||D.customElements.define(`media-mute-button`,Cs);var ws=Cs,Ts=`<svg aria-hidden="true" viewBox="0 0 28 24">
  <path d="M24 3H4a1 1 0 0 0-1 1v16a1 1 0 0 0 1 1h20a1 1 0 0 0 1-1V4a1 1 0 0 0-1-1Zm-1 16H5V5h18v14Zm-3-8h-7v5h7v-5Z"/>
</svg>`;function Es(e){return`
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
      <slot name="enter">${Ts}</slot>
      <slot name="exit">${Ts}</slot>
    </slot>
  `}function Ds(){return`
    <slot name="tooltip-enter">${E(`Enter picture in picture mode`)}</slot>
    <slot name="tooltip-exit">${E(`Exit picture in picture mode`)}</slot>
  `}var Os=e=>{let t=e.mediaIsPip?E(`exit picture in picture mode`):E(`enter picture in picture mode`);e.setAttribute(`aria-label`,t)},ks=class extends G{static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_IS_PIP,b.MEDIA_PIP_UNAVAILABLE]}connectedCallback(){super.connectedCallback(),Os(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===b.MEDIA_IS_PIP&&Os(this)}get mediaPipUnavailable(){return P(this,b.MEDIA_PIP_UNAVAILABLE)}set mediaPipUnavailable(e){F(this,b.MEDIA_PIP_UNAVAILABLE,e)}get mediaIsPip(){return M(this,b.MEDIA_IS_PIP)}set mediaIsPip(e){N(this,b.MEDIA_IS_PIP,e)}handleClick(){let e=this.mediaIsPip?v.MEDIA_EXIT_PIP_REQUEST:v.MEDIA_ENTER_PIP_REQUEST;this.dispatchEvent(new D.CustomEvent(e,{composed:!0,bubbles:!0}))}};ks.getSlotTemplateHTML=Es,ks.getTooltipContentHTML=Ds,D.customElements.get(`media-pip-button`)||D.customElements.define(`media-pip-button`,ks);var As=ks,js=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Ms=(e,t,n)=>(js(e,t,`read from private field`),n?n.call(e):t.get(e)),Ns=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Ps,Fs={RATES:`rates`},Is=[1,1.2,1.5,1.7,2];function Ls(e){return Math.round(e*100)/100}function Rs(e){return`
    <style>
      :host {
        min-width: 5ch;
        padding: var(--media-button-padding, var(--media-control-padding, 10px 5px));
      }
    </style>
    <slot name="icon">${e.mediaplaybackrate?Ls(+e.mediaplaybackrate):1}x</slot>
  `}function zs(){return E(`Playback rate`)}var Bs=class extends G{constructor(){super(),Ns(this,Ps,new Sn(this,Fs.RATES,{defaultValue:Is})),this.container=this.shadowRoot.querySelector(`slot[name="icon"]`),this.container.innerHTML=`${Ls(this.mediaPlaybackRate??1)}x`}static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_PLAYBACK_RATE,Fs.RATES]}attributeChangedCallback(e,t,n){if(super.attributeChangedCallback(e,t,n),e===Fs.RATES&&(Ms(this,Ps).value=n),e===b.MEDIA_PLAYBACK_RATE){let e=n?+n:NaN,t=Ls(Number.isNaN(e)?1:e);this.container.innerHTML=`${t}x`,this.setAttribute(`aria-label`,E(`Playback rate {playbackRate}`,{playbackRate:t}))}}get rates(){return Ms(this,Ps)}set rates(e){e?Array.isArray(e)?Ms(this,Ps).value=e.join(` `):typeof e==`string`&&(Ms(this,Ps).value=e):Ms(this,Ps).value=``}get mediaPlaybackRate(){return A(this,b.MEDIA_PLAYBACK_RATE,1)}set mediaPlaybackRate(e){j(this,b.MEDIA_PLAYBACK_RATE,e)}handleClick(){let e=Array.from(Ms(this,Ps).values(),e=>+e).sort((e,t)=>e-t),t=e.find(e=>e>this.mediaPlaybackRate)??e[0]??1,n=new D.CustomEvent(v.MEDIA_PLAYBACK_RATE_REQUEST,{composed:!0,bubbles:!0,detail:t});this.dispatchEvent(n)}};Ps=new WeakMap,Bs.getSlotTemplateHTML=Rs,Bs.getTooltipContentHTML=zs,D.customElements.get(`media-playback-rate-button`)||D.customElements.define(`media-playback-rate-button`,Bs);var Vs=Bs,Hs=`<svg aria-hidden="true" viewBox="0 0 24 24">
  <path d="m6 21 15-9L6 3v18Z"/>
</svg>`,Us=`<svg aria-hidden="true" viewBox="0 0 24 24">
  <path d="M6 20h4V4H6v16Zm8-16v16h4V4h-4Z"/>
</svg>`;function Ws(e){return`
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
      <slot name="play">${Hs}</slot>
      <slot name="pause">${Us}</slot>
    </slot>
  `}function Gs(){return`
    <slot name="tooltip-play">${E(`Play`)}</slot>
    <slot name="tooltip-pause">${E(`Pause`)}</slot>
  `}var Ks=e=>{let t=e.mediaPaused?E(`play`):E(`pause`);e.setAttribute(`aria-label`,t)},qs=class extends G{static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_PAUSED,b.MEDIA_ENDED]}connectedCallback(){super.connectedCallback(),Ks(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),(e===b.MEDIA_PAUSED||e===b.MEDIA_LANG)&&Ks(this)}get mediaPaused(){return M(this,b.MEDIA_PAUSED)}set mediaPaused(e){N(this,b.MEDIA_PAUSED,e)}handleClick(){let e=this.mediaPaused?v.MEDIA_PLAY_REQUEST:v.MEDIA_PAUSE_REQUEST;this.dispatchEvent(new D.CustomEvent(e,{composed:!0,bubbles:!0}))}};qs.getSlotTemplateHTML=Ws,qs.getTooltipContentHTML=Gs,D.customElements.get(`media-play-button`)||D.customElements.define(`media-play-button`,qs);var Js=qs,Ys={PLACEHOLDER_SRC:`placeholdersrc`,SRC:`src`};function Xs(e){return`
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
  `}var Zs=e=>{e.style.removeProperty(`background-image`)},Qs=(e,t)=>{e.style[`background-image`]=`url('${t}')`},$s=class extends D.HTMLElement{static get observedAttributes(){return[Ys.PLACEHOLDER_SRC,Ys.SRC]}constructor(){if(super(),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=ft(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}this.image=this.shadowRoot.querySelector(`#image`)}attributeChangedCallback(e,t,n){e===Ys.SRC&&(n==null?this.image.removeAttribute(Ys.SRC):this.image.setAttribute(Ys.SRC,n)),e===Ys.PLACEHOLDER_SRC&&(n==null?Zs(this.image):Qs(this.image,n))}get placeholderSrc(){return P(this,Ys.PLACEHOLDER_SRC)}set placeholderSrc(e){F(this,Ys.SRC,e)}get src(){return P(this,Ys.SRC)}set src(e){F(this,Ys.SRC,e)}};$s.shadowRootOptions={mode:`open`},$s.getTemplateHTML=Xs,D.customElements.get(`media-poster-image`)||D.customElements.define(`media-poster-image`,$s);var ec=$s,tc=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},nc=(e,t,n)=>(tc(e,t,`read from private field`),n?n.call(e):t.get(e)),rc=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},ic=(e,t,n,r)=>(tc(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),ac,oc=class extends io{constructor(){super(),rc(this,ac,void 0),ic(this,ac,this.shadowRoot.querySelector(`slot`))}static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_PREVIEW_CHAPTER,b.MEDIA_LANG]}attributeChangedCallback(e,t,n){if(super.attributeChangedCallback(e,t,n),(e===b.MEDIA_PREVIEW_CHAPTER||e===b.MEDIA_LANG)&&n!==t&&n!=null){if(nc(this,ac).textContent=n,n!==``){let e=E(`chapter: {chapterName}`,{chapterName:n});this.setAttribute(`aria-valuetext`,e)}else this.removeAttribute(`aria-valuetext`)}}get mediaPreviewChapter(){return P(this,b.MEDIA_PREVIEW_CHAPTER)}set mediaPreviewChapter(e){F(this,b.MEDIA_PREVIEW_CHAPTER,e)}};ac=new WeakMap,D.customElements.get(`media-preview-chapter-display`)||D.customElements.define(`media-preview-chapter-display`,oc);var sc=oc,cc=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},lc=(e,t,n)=>(cc(e,t,`read from private field`),n?n.call(e):t.get(e)),uc=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},dc=(e,t,n,r)=>(cc(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),fc;function pc(e){return`
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
  `}var mc=class extends D.HTMLElement{constructor(){if(super(),uc(this,fc,void 0),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=ft(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}}static get observedAttributes(){return[y.MEDIA_CONTROLLER,b.MEDIA_PREVIEW_IMAGE,b.MEDIA_PREVIEW_COORDS]}connectedCallback(){var e,t;let n=this.getAttribute(y.MEDIA_CONTROLLER);n&&(dc(this,fc,this.getRootNode()?.getElementById(n)),(t=(e=lc(this,fc))?.associateElement)==null||t.call(e,this))}disconnectedCallback(){var e,t;(t=(e=lc(this,fc))?.unassociateElement)==null||t.call(e,this),dc(this,fc,null)}attributeChangedCallback(e,t,n){var r,i,a,o;[b.MEDIA_PREVIEW_IMAGE,b.MEDIA_PREVIEW_COORDS].includes(e)&&this.update(),e===y.MEDIA_CONTROLLER&&(t&&((i=(r=lc(this,fc))?.unassociateElement)==null||i.call(r,this),dc(this,fc,null)),n&&this.isConnected&&(dc(this,fc,this.getRootNode()?.getElementById(n)),(o=(a=lc(this,fc))?.associateElement)==null||o.call(a,this)))}get mediaPreviewImage(){return P(this,b.MEDIA_PREVIEW_IMAGE)}set mediaPreviewImage(e){F(this,b.MEDIA_PREVIEW_IMAGE,e)}get mediaPreviewCoords(){let e=this.getAttribute(b.MEDIA_PREVIEW_COORDS);if(e)return e.split(/\s+/).map(e=>+e)}set mediaPreviewCoords(e){if(!e){this.removeAttribute(b.MEDIA_PREVIEW_COORDS);return}this.setAttribute(b.MEDIA_PREVIEW_COORDS,e.join(` `))}update(){let e=this.mediaPreviewCoords,t=this.mediaPreviewImage;if(!(e&&t))return;let[n,r,i,a]=e,o=t.split(`#`)[0],s=getComputedStyle(this),{maxWidth:c,maxHeight:l,minWidth:u,minHeight:d}=s,f=s.getPropertyValue(`--media-preview-thumbnail-object-fit`).trim()||`contain`,p,m;if(f===`fill`){let e=parseInt(c)/i,t=parseInt(l)/a,n=parseInt(u)/i,r=parseInt(d)/a;p=e<1?e:Math.max(e,n),m=t<1?t:Math.max(t,r)}else{let e=Math.min(parseInt(c)/i,parseInt(l)/a),t=Math.max(parseInt(u)/i,parseInt(d)/a),n=e<1?e:t>1?t:1;p=n,m=n}let{style:h}=k(this.shadowRoot,`:host`),ee=k(this.shadowRoot,`img`).style,te=this.shadowRoot.querySelector(`img`),ne=Math.min(p,m)<1?`min`:`max`;h.setProperty(`${ne}-width`,`initial`,`important`),h.setProperty(`${ne}-height`,`initial`,`important`),h.width=`${i*p}px`,h.height=`${a*m}px`;let re=()=>{ee.width=`${this.imgWidth*p}px`,ee.height=`${this.imgHeight*m}px`,ee.display=`block`};te.src!==o&&(te.onload=()=>{this.imgWidth=te.naturalWidth,this.imgHeight=te.naturalHeight,re(),te.onload=null},te.src=o,re()),re(),ee.transform=`translate(-${n*p}px, -${r*m}px)`}};fc=new WeakMap,mc.shadowRootOptions={mode:`open`},mc.getTemplateHTML=pc,D.customElements.get(`media-preview-thumbnail`)||D.customElements.define(`media-preview-thumbnail`,mc);var hc=mc,gc=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},_c=(e,t,n)=>(gc(e,t,`read from private field`),n?n.call(e):t.get(e)),vc=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},yc=(e,t,n,r)=>(gc(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),bc,xc=class extends io{constructor(){super(),vc(this,bc,void 0),yc(this,bc,this.shadowRoot.querySelector(`slot`)),_c(this,bc).textContent=Qe(0)}static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_PREVIEW_TIME]}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===b.MEDIA_PREVIEW_TIME&&n!=null&&(_c(this,bc).textContent=Qe(parseFloat(n)))}get mediaPreviewTime(){return A(this,b.MEDIA_PREVIEW_TIME)}set mediaPreviewTime(e){j(this,b.MEDIA_PREVIEW_TIME,e)}};bc=new WeakMap,D.customElements.get(`media-preview-time-display`)||D.customElements.define(`media-preview-time-display`,xc);var Sc=xc,Cc={SEEK_OFFSET:`seekoffset`},wc=30,Tc=e=>`
  <svg aria-hidden="true" viewBox="0 0 20 24">
    <defs>
      <style>.text{font-size:8px;font-family:Arial-BoldMT, Arial;font-weight:700;}</style>
    </defs>
    <text class="text value" transform="translate(2.18 19.87)">${e}</text>
    <path d="M10 6V3L4.37 7 10 10.94V8a5.54 5.54 0 0 1 1.9 10.48v2.12A7.5 7.5 0 0 0 10 6Z"/>
  </svg>`;function Ec(e,t){return`
    <slot name="icon">${Tc(t.seekOffset)}</slot>
  `}var Dc=(e,t)=>{e.setAttribute(`aria-label`,E(`seek back {seekOffset} seconds`,{seekOffset:t}))};function Oc(){return E(`Seek backward`)}var kc=0,Ac=class extends G{static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_CURRENT_TIME,Cc.SEEK_OFFSET]}connectedCallback(){super.connectedCallback(),this.seekOffset=A(this,Cc.SEEK_OFFSET,wc)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),Dc(this,this.seekOffset),e===Cc.SEEK_OFFSET&&(this.seekOffset=A(this,Cc.SEEK_OFFSET,wc))}get seekOffset(){return A(this,Cc.SEEK_OFFSET,wc)}set seekOffset(e){j(this,Cc.SEEK_OFFSET,e),this.setAttribute(`aria-label`,E(`seek back {seekOffset} seconds`,{seekOffset:this.seekOffset})),ht(_t(this,`icon`),this.seekOffset)}get mediaCurrentTime(){return A(this,b.MEDIA_CURRENT_TIME,kc)}set mediaCurrentTime(e){j(this,b.MEDIA_CURRENT_TIME,e)}handleClick(){let e=Math.max(this.mediaCurrentTime-this.seekOffset,0),t=new D.CustomEvent(v.MEDIA_SEEK_REQUEST,{composed:!0,bubbles:!0,detail:e});this.dispatchEvent(t)}};Ac.getSlotTemplateHTML=Ec,Ac.getTooltipContentHTML=Oc,D.customElements.get(`media-seek-backward-button`)||D.customElements.define(`media-seek-backward-button`,Ac);var jc=Ac,Mc={SEEK_OFFSET:`seekoffset`},Nc=30,Pc=e=>`
  <svg aria-hidden="true" viewBox="0 0 20 24">
    <defs>
      <style>.text{font-size:8px;font-family:Arial-BoldMT, Arial;font-weight:700;}</style>
    </defs>
    <text class="text value" transform="translate(8.9 19.87)">${e}</text>
    <path d="M10 6V3l5.61 4L10 10.94V8a5.54 5.54 0 0 0-1.9 10.48v2.12A7.5 7.5 0 0 1 10 6Z"/>
  </svg>`;function Fc(e,t){return`
    <slot name="icon">${Pc(t.seekOffset)}</slot>
  `}var Ic=(e,t)=>{e.setAttribute(`aria-label`,E(`seek forward {seekOffset} seconds`,{seekOffset:t}))};function Lc(){return E(`Seek forward`)}var Rc=0,zc=class extends G{static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_CURRENT_TIME,Mc.SEEK_OFFSET]}connectedCallback(){super.connectedCallback(),this.seekOffset=A(this,Mc.SEEK_OFFSET,Nc)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),Ic(this,this.seekOffset),e===Mc.SEEK_OFFSET&&(this.seekOffset=A(this,Mc.SEEK_OFFSET,Nc))}get seekOffset(){return A(this,Mc.SEEK_OFFSET,Nc)}set seekOffset(e){j(this,Mc.SEEK_OFFSET,e),this.setAttribute(`aria-label`,E(`seek forward {seekOffset} seconds`,{seekOffset:this.seekOffset})),ht(_t(this,`icon`),this.seekOffset)}get mediaCurrentTime(){return A(this,b.MEDIA_CURRENT_TIME,Rc)}set mediaCurrentTime(e){j(this,b.MEDIA_CURRENT_TIME,e)}handleClick(){let e=this.mediaCurrentTime+this.seekOffset,t=new D.CustomEvent(v.MEDIA_SEEK_REQUEST,{composed:!0,bubbles:!0,detail:e});this.dispatchEvent(t)}};zc.getSlotTemplateHTML=Fc,zc.getTooltipContentHTML=Lc,D.customElements.get(`media-seek-forward-button`)||D.customElements.define(`media-seek-forward-button`,zc);var Bc=zc,Vc=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Hc=(e,t,n)=>(Vc(e,t,`read from private field`),n?n.call(e):t.get(e)),Uc=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Wc=(e,t,n,r)=>(Vc(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),Gc=(e,t,n)=>(Vc(e,t,`access private method`),n),Kc,qc,Jc,Yc,Xc,Zc,Qc,$c,el,tl,nl,rl={REMAINING:`remaining`,SHOW_DURATION:`showduration`,NO_TOGGLE:`notoggle`},il=[...Object.values(rl),b.MEDIA_CURRENT_TIME,b.MEDIA_DURATION,b.MEDIA_SEEKABLE],al=[`Enter`,` `],ol=`&nbsp;/&nbsp;`,sl=(e,{timesSep:t=ol}={})=>{let n=e.mediaCurrentTime??0,[,r]=e.mediaSeekable??[],i=0;Number.isFinite(e.mediaDuration)?i=e.mediaDuration:Number.isFinite(r)&&(i=r);let a=e.remaining?Qe(0-(i-n)):Qe(n);return e.showDuration?`${a}${t}${Qe(i)}`:a},cl=e=>{let t=e.mediaCurrentTime,[,n]=e.mediaSeekable??[],r=null;if(Number.isFinite(e.mediaDuration)?r=e.mediaDuration:Number.isFinite(n)&&(r=n),t==null||r===null){e.setAttribute(`aria-description`,E(`video not loaded, unknown time.`));return}let i=e.remaining?Ze(0-(r-t)):Ze(t);if(!e.showDuration){e.setAttribute(`aria-description`,i);return}let a=E(`{currentTime} of {totalTime}`,{currentTime:i,totalTime:Ze(r)});e.setAttribute(`aria-description`,a)};function ll(e,t){return`
    <slot>${sl(t)}</slot>
  `}var ul=e=>{e.setAttribute(`aria-label`,E(`playback time`))},dl=class extends io{constructor(){super(),Uc(this,Yc),Uc(this,Zc),Uc(this,$c),Uc(this,tl),Uc(this,Kc,void 0),Uc(this,qc,null),Uc(this,Jc,e=>{let{metaKey:t,altKey:n,key:r}=e;if(t||n||!al.includes(r)){this.removeEventListener(`keyup`,Hc(this,qc));return}this.addEventListener(`keyup`,Hc(this,qc))}),Wc(this,Kc,this.shadowRoot.querySelector(`slot`)),Hc(this,Kc).innerHTML=`${sl(this)}`}static get observedAttributes(){return[...super.observedAttributes,...il,`disabled`]}connectedCallback(){let{style:e}=k(this.shadowRoot,`:host(:hover:not([notoggle]))`);e.setProperty(`cursor`,`var(--media-cursor, pointer)`),e.setProperty(`background`,`var(--media-control-hover-background, rgba(50 50 70 / .7))`),this.setAttribute(`aria-label`,E(`playback time`)),Gc(this,$c,el).call(this),super.connectedCallback()}toggleTimeDisplay(){this.noToggle||(this.hasAttribute(`remaining`)?this.removeAttribute(`remaining`):this.setAttribute(`remaining`,``))}disconnectedCallback(){this.disable(),Gc(this,Zc,Qc).call(this),super.disconnectedCallback()}attributeChangedCallback(e,t,n){ul(this),il.includes(e)?this.update():e===`disabled`&&n!==t?n==null?Gc(this,$c,el).call(this):Gc(this,tl,nl).call(this):e===rl.NO_TOGGLE&&n!==t&&(this.noToggle?Gc(this,tl,nl).call(this):Gc(this,$c,el).call(this)),super.attributeChangedCallback(e,t,n)}enable(){this.noToggle||(this.tabIndex=0)}disable(){this.tabIndex=-1}get remaining(){return M(this,rl.REMAINING)}set remaining(e){N(this,rl.REMAINING,e)}get showDuration(){return M(this,rl.SHOW_DURATION)}set showDuration(e){N(this,rl.SHOW_DURATION,e)}get noToggle(){return M(this,rl.NO_TOGGLE)}set noToggle(e){N(this,rl.NO_TOGGLE,e)}get mediaDuration(){return A(this,b.MEDIA_DURATION)}set mediaDuration(e){j(this,b.MEDIA_DURATION,e)}get mediaCurrentTime(){return A(this,b.MEDIA_CURRENT_TIME)}set mediaCurrentTime(e){j(this,b.MEDIA_CURRENT_TIME,e)}get mediaSeekable(){let e=this.getAttribute(b.MEDIA_SEEKABLE);if(e)return e.split(`:`).map(e=>+e)}set mediaSeekable(e){if(e==null){this.removeAttribute(b.MEDIA_SEEKABLE);return}this.setAttribute(b.MEDIA_SEEKABLE,e.join(`:`))}update(){let e=sl(this);cl(this),e!==Hc(this,Kc).innerHTML&&(Hc(this,Kc).innerHTML=e)}};Kc=new WeakMap,qc=new WeakMap,Jc=new WeakMap,Yc=new WeakSet,Xc=function(){Hc(this,qc)||(Wc(this,qc,e=>{let{key:t}=e;if(!al.includes(t)){this.removeEventListener(`keyup`,Hc(this,qc));return}this.toggleTimeDisplay()}),this.addEventListener(`keydown`,Hc(this,Jc)),this.addEventListener(`click`,this.toggleTimeDisplay))},Zc=new WeakSet,Qc=function(){Hc(this,qc)&&(this.removeEventListener(`keyup`,Hc(this,qc)),this.removeEventListener(`keydown`,Hc(this,Jc)),this.removeEventListener(`click`,this.toggleTimeDisplay),Wc(this,qc,null))},$c=new WeakSet,el=function(){!this.noToggle&&!this.hasAttribute(`disabled`)&&(this.setAttribute(`role`,`button`),this.enable(),Gc(this,Yc,Xc).call(this))},tl=new WeakSet,nl=function(){this.removeAttribute(`role`),this.disable(),Gc(this,Zc,Qc).call(this)},dl.getSlotTemplateHTML=ll,D.customElements.get(`media-time-display`)||D.customElements.define(`media-time-display`,dl);var fl=dl,pl=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},J=(e,t,n)=>(pl(e,t,`read from private field`),n?n.call(e):t.get(e)),ml=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Y=(e,t,n,r)=>(pl(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),hl=(e,t,n,r)=>({set _(r){Y(e,t,r,n)},get _(){return J(e,t,r)}}),gl,_l,vl,yl,bl,xl,Sl,Cl,wl,Tl,El=class{constructor(e,t,n){ml(this,gl,void 0),ml(this,_l,void 0),ml(this,vl,void 0),ml(this,yl,void 0),ml(this,bl,void 0),ml(this,xl,void 0),ml(this,Sl,void 0),ml(this,Cl,void 0),ml(this,wl,0),ml(this,Tl,(e=performance.now())=>{Y(this,wl,requestAnimationFrame(J(this,Tl))),Y(this,yl,performance.now()-J(this,vl));let t=1e3/this.fps;if(J(this,yl)>t){Y(this,vl,e-J(this,yl)%t);let n=1e3/((e-J(this,_l))/++hl(this,bl)._),r=(e-J(this,xl))/1e3/this.duration,i=J(this,Sl)+r*this.playbackRate;i-J(this,gl).valueAsNumber>0?Y(this,Cl,this.playbackRate/this.duration/n):(Y(this,Cl,.995*J(this,Cl)),i=J(this,gl).valueAsNumber+J(this,Cl)),this.callback(i)}}),Y(this,gl,e),this.callback=t,this.fps=n}start(){J(this,wl)===0&&(Y(this,vl,performance.now()),Y(this,_l,J(this,vl)),Y(this,bl,0),J(this,Tl).call(this))}stop(){J(this,wl)!==0&&(cancelAnimationFrame(J(this,wl)),Y(this,wl,0))}update({start:e,duration:t,playbackRate:n}){let r=e-J(this,gl).valueAsNumber,i=Math.abs(t-this.duration);(r>0||r<-.5/t||i>=.5)&&this.callback(e),Y(this,Sl,e),Y(this,xl,performance.now()),this.duration=t,this.playbackRate=n}};gl=new WeakMap,_l=new WeakMap,vl=new WeakMap,yl=new WeakMap,bl=new WeakMap,xl=new WeakMap,Sl=new WeakMap,Cl=new WeakMap,wl=new WeakMap,Tl=new WeakMap;var Dl=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},X=(e,t,n)=>(Dl(e,t,`read from private field`),n?n.call(e):t.get(e)),Z=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Ol=(e,t,n,r)=>(Dl(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),kl=(e,t,n)=>(Dl(e,t,`access private method`),n),Al,jl,Ml,Nl,Pl,Fl,Il,Ll,Rl,zl,Bl,Vl,Hl,Ul,Wl,Gl,Kl,ql,Jl,Yl,Xl,Zl,Ql,$l,eu,tu,nu=e=>{let t=e.range,n=Ze(+au(e)),r=Ze(+e.mediaSeekableEnd),i=n&&r?E(`{currentTime} of {totalTime}`,{currentTime:n,totalTime:r}):E(`video not loaded, unknown time.`);t.setAttribute(`aria-valuetext`,i)};function ru(e){return`
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
          <template shadowrootmode="${hc.shadowRootOptions.mode}">
            ${hc.getTemplateHTML({})}
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
  `}var iu=(e,t=e.mediaCurrentTime)=>{let n=Number.isFinite(e.mediaSeekableStart)?e.mediaSeekableStart:0,r=Number.isFinite(e.mediaDuration)?e.mediaDuration:e.mediaSeekableEnd;if(Number.isNaN(r))return 0;let i=(t-n)/(r-n);return Math.max(0,Math.min(i,1))},au=(e,t=e.range.valueAsNumber)=>{let n=Number.isFinite(e.mediaSeekableStart)?e.mediaSeekableStart:0,r=Number.isFinite(e.mediaDuration)?e.mediaDuration:e.mediaSeekableEnd;return Number.isNaN(r)?0:t*(r-n)+n},ou=class extends Va{constructor(){super(),Z(this,Vl),Z(this,Wl),Z(this,Kl),Z(this,Jl),Z(this,Xl),Z(this,Ql),Z(this,eu),Z(this,Al,null),Z(this,jl,void 0),Z(this,Ml,void 0),Z(this,Nl,void 0),Z(this,Pl,void 0),Z(this,Fl,void 0),Z(this,Il,void 0),Z(this,Ll,void 0),Z(this,Rl,void 0),Z(this,zl,void 0),Z(this,Bl,()=>{kl(this,Vl,Hl).call(this)?X(this,jl).start():X(this,jl).stop()}),Z(this,Ul,e=>{this.dragging||(He(e)&&(this.range.valueAsNumber=e),X(this,zl)||this.updateBar())}),this.shadowRoot.querySelector(`#track`).insertAdjacentHTML(`afterbegin`,`<div id="buffered" part="buffered"></div>`),Ol(this,Ml,this.shadowRoot.querySelectorAll(`[part~="box"]`)),Ol(this,Pl,this.shadowRoot.querySelector(`[part~="preview-box"]`)),Ol(this,Fl,this.shadowRoot.querySelector(`[part~="current-box"]`));let e=getComputedStyle(this);Ol(this,Il,parseInt(e.getPropertyValue(`--media-box-padding-left`))),Ol(this,Ll,parseInt(e.getPropertyValue(`--media-box-padding-right`))),Ol(this,jl,new El(this.range,X(this,Ul),60))}static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_PAUSED,b.MEDIA_DURATION,b.MEDIA_SEEKABLE,b.MEDIA_CURRENT_TIME,b.MEDIA_PREVIEW_IMAGE,b.MEDIA_PREVIEW_TIME,b.MEDIA_PREVIEW_CHAPTER,b.MEDIA_BUFFERED,b.MEDIA_PLAYBACK_RATE,b.MEDIA_LOADING,b.MEDIA_ENDED]}connectedCallback(){var e;super.connectedCallback(),this.range.setAttribute(`aria-label`,E(`seek`)),X(this,Bl).call(this),Ol(this,Al,this.getRootNode()),(e=X(this,Al))==null||e.addEventListener(`transitionstart`,this)}disconnectedCallback(){var e;super.disconnectedCallback(),X(this,jl).stop(),(e=X(this,Al))==null||e.removeEventListener(`transitionstart`,this),Ol(this,Al,null)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),t!=n&&(e===b.MEDIA_CURRENT_TIME||e===b.MEDIA_PAUSED||e===b.MEDIA_ENDED||e===b.MEDIA_LOADING||e===b.MEDIA_DURATION||e===b.MEDIA_SEEKABLE?(X(this,jl).update({start:iu(this),duration:this.mediaSeekableEnd-this.mediaSeekableStart,playbackRate:this.mediaPlaybackRate}),X(this,Bl).call(this),nu(this)):e===b.MEDIA_BUFFERED&&this.updateBufferedBar(),(e===b.MEDIA_DURATION||e===b.MEDIA_SEEKABLE)&&(this.mediaChaptersCues=X(this,Rl),this.updateBar()))}get mediaChaptersCues(){return X(this,Rl)}set mediaChaptersCues(e){Ol(this,Rl,e),this.updateSegments(X(this,Rl)?.map(e=>({start:iu(this,e.startTime),end:iu(this,e.endTime)})))}get mediaPaused(){return M(this,b.MEDIA_PAUSED)}set mediaPaused(e){N(this,b.MEDIA_PAUSED,e)}get mediaLoading(){return M(this,b.MEDIA_LOADING)}set mediaLoading(e){N(this,b.MEDIA_LOADING,e)}get mediaDuration(){return A(this,b.MEDIA_DURATION)}set mediaDuration(e){j(this,b.MEDIA_DURATION,e)}get mediaCurrentTime(){return A(this,b.MEDIA_CURRENT_TIME)}set mediaCurrentTime(e){j(this,b.MEDIA_CURRENT_TIME,e)}get mediaPlaybackRate(){return A(this,b.MEDIA_PLAYBACK_RATE,1)}set mediaPlaybackRate(e){j(this,b.MEDIA_PLAYBACK_RATE,e)}get mediaBuffered(){let e=this.getAttribute(b.MEDIA_BUFFERED);return e?e.split(` `).map(e=>e.split(`:`).map(e=>+e)):[]}set mediaBuffered(e){if(!e){this.removeAttribute(b.MEDIA_BUFFERED);return}let t=e.map(e=>e.join(`:`)).join(` `);this.setAttribute(b.MEDIA_BUFFERED,t)}get mediaSeekable(){let e=this.getAttribute(b.MEDIA_SEEKABLE);if(e)return e.split(`:`).map(e=>+e)}set mediaSeekable(e){if(e==null){this.removeAttribute(b.MEDIA_SEEKABLE);return}this.setAttribute(b.MEDIA_SEEKABLE,e.join(`:`))}get mediaSeekableEnd(){let[,e=this.mediaDuration]=this.mediaSeekable??[];return e}get mediaSeekableStart(){let[e=0]=this.mediaSeekable??[];return e}get mediaPreviewImage(){return P(this,b.MEDIA_PREVIEW_IMAGE)}set mediaPreviewImage(e){F(this,b.MEDIA_PREVIEW_IMAGE,e)}get mediaPreviewTime(){return A(this,b.MEDIA_PREVIEW_TIME)}set mediaPreviewTime(e){j(this,b.MEDIA_PREVIEW_TIME,e)}get mediaEnded(){return M(this,b.MEDIA_ENDED)}set mediaEnded(e){N(this,b.MEDIA_ENDED,e)}updateBar(){super.updateBar(),this.updateBufferedBar(),this.updateCurrentBox()}updateBufferedBar(){let e=this.mediaBuffered;if(!e.length)return;let t;if(this.mediaEnded)t=1;else{let n=this.mediaCurrentTime,[,r=this.mediaSeekableStart]=e.find(([e,t])=>e<=n&&n<=t)??[];t=iu(this,r)}let{style:n}=k(this.shadowRoot,`#buffered`);n.setProperty(`width`,`${t*100}%`)}updateCurrentBox(){if(!this.shadowRoot.querySelector(`slot[name="current"]`).assignedElements().length)return;let e=k(this.shadowRoot,`#current-rail`),t=k(this.shadowRoot,`[part~="current-box"]`),n=kl(this,Wl,Gl).call(this,X(this,Fl)),r=kl(this,Kl,ql).call(this,n,this.range.valueAsNumber),i=kl(this,Jl,Yl).call(this,n,this.range.valueAsNumber);e.style.transform=`translateX(${r})`,e.style.setProperty(`--_range-width`,`${n.range.width}`),t.style.setProperty(`--_box-shift`,`${i}`),t.style.setProperty(`--_box-width`,`${n.box.width}px`),t.style.setProperty(`visibility`,`initial`)}handleEvent(e){switch(super.handleEvent(e),e.type){case`input`:kl(this,eu,tu).call(this);break;case`pointermove`:kl(this,Xl,Zl).call(this,e);break;case`pointerup`:X(this,zl)&&Ol(this,zl,!1);break;case`pointerdown`:Ol(this,zl,!0);break;case`pointerleave`:kl(this,Ql,$l).call(this,null);break;case`transitionstart`:vt(e.target,this)&&setTimeout(()=>X(this,Bl).call(this),0)}}};Al=new WeakMap,jl=new WeakMap,Ml=new WeakMap,Nl=new WeakMap,Pl=new WeakMap,Fl=new WeakMap,Il=new WeakMap,Ll=new WeakMap,Rl=new WeakMap,zl=new WeakMap,Bl=new WeakMap,Vl=new WeakSet,Hl=function(){return this.isConnected&&!this.mediaPaused&&!this.mediaLoading&&!this.mediaEnded&&this.mediaSeekableEnd>0&&St(this)},Ul=new WeakMap,Wl=new WeakSet,Gl=function(e){let t=((this.getAttribute(`bounds`)?yt(this,`#${this.getAttribute(`bounds`)}`):this.parentElement)??this).getBoundingClientRect(),n=this.range.getBoundingClientRect(),r=e.offsetWidth;return{box:{width:r,min:-(n.left-t.left-r/2),max:t.right-n.left-r/2},bounds:t,range:n}},Kl=new WeakSet,ql=function(e,t){let n=`${t*100}%`,{width:r,min:i,max:a}=e.box;if(!r)return n;if(Number.isNaN(i)||(n=`max(${`calc(1 / var(--_range-width) * 100 * ${i}% + var(--media-box-padding-left))`}, ${n})`),!Number.isNaN(a)){let e=`calc(1 / var(--_range-width) * 100 * ${a}% - var(--media-box-padding-right))`;n=`min(${n}, ${e})`}return n},Jl=new WeakSet,Yl=function(e,t){let{width:n,min:r,max:i}=e.box,a=t*e.range.width;if(a<r+X(this,Il)){let t=e.range.left-e.bounds.left-X(this,Il);return`${a-n/2+t}px`}if(a>i-X(this,Ll)){let t=e.bounds.right-e.range.right-X(this,Ll);return`${a+n/2-t-e.range.width}px`}return 0},Xl=new WeakSet,Zl=function(e){let t=[...X(this,Ml)].some(t=>e.composedPath().includes(t));if(!this.dragging&&(t||!e.composedPath().includes(this))){kl(this,Ql,$l).call(this,null);return}let n=this.mediaSeekableEnd;if(!n)return;let r=k(this.shadowRoot,`#preview-rail`),i=k(this.shadowRoot,`[part~="preview-box"]`),a=kl(this,Wl,Gl).call(this,X(this,Pl)),o=(e.clientX-a.range.left)/a.range.width;o=Math.max(0,Math.min(1,o));let s=kl(this,Kl,ql).call(this,a,o),c=kl(this,Jl,Yl).call(this,a,o);r.style.transform=`translateX(${s})`,r.style.setProperty(`--_range-width`,`${a.range.width}`),i.style.setProperty(`--_box-shift`,`${c}`),i.style.setProperty(`--_box-width`,`${a.box.width}px`);let l=Math.round(X(this,Nl))-Math.round(o*n);Math.abs(l)<1&&o>.01&&o<.99||(Ol(this,Nl,o*n),kl(this,Ql,$l).call(this,X(this,Nl)))},Ql=new WeakSet,$l=function(e){this.dispatchEvent(new D.CustomEvent(v.MEDIA_PREVIEW_REQUEST,{composed:!0,bubbles:!0,detail:e}))},eu=new WeakSet,tu=function(){X(this,jl).stop();let e=au(this);this.dispatchEvent(new D.CustomEvent(v.MEDIA_SEEK_REQUEST,{composed:!0,bubbles:!0,detail:e}))},ou.shadowRootOptions={mode:`open`},ou.getContainerTemplateHTML=ru,D.customElements.get(`media-time-range`)||D.customElements.define(`media-time-range`,ou);var su=ou,cu=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},lu=(e,t,n)=>(cu(e,t,`read from private field`),n?n.call(e):t.get(e)),uu=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},du,fu=1,pu=e=>e.mediaMuted?0:e.mediaVolume,mu=e=>`${Math.round(e*100)}%`,hu=class extends Va{constructor(){super(...arguments),uu(this,du,()=>{let e=this.range.value,t=new D.CustomEvent(v.MEDIA_VOLUME_REQUEST,{composed:!0,bubbles:!0,detail:e});this.dispatchEvent(t)})}static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_VOLUME,b.MEDIA_MUTED,b.MEDIA_VOLUME_UNAVAILABLE]}connectedCallback(){super.connectedCallback(),this.range.setAttribute(`aria-label`,E(`volume`)),this.range.addEventListener(`input`,lu(this,du))}disconnectedCallback(){this.range.removeEventListener(`input`,lu(this,du)),super.disconnectedCallback()}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),(e===b.MEDIA_VOLUME||e===b.MEDIA_MUTED)&&(this.range.valueAsNumber=pu(this),this.range.setAttribute(`aria-valuetext`,mu(this.range.valueAsNumber)),this.updateBar())}get mediaVolume(){return A(this,b.MEDIA_VOLUME,fu)}set mediaVolume(e){j(this,b.MEDIA_VOLUME,e)}get mediaMuted(){return M(this,b.MEDIA_MUTED)}set mediaMuted(e){N(this,b.MEDIA_MUTED,e)}get mediaVolumeUnavailable(){return P(this,b.MEDIA_VOLUME_UNAVAILABLE)}set mediaVolumeUnavailable(e){F(this,b.MEDIA_VOLUME_UNAVAILABLE,e)}};du=new WeakMap,D.customElements.get(`media-volume-range`)||D.customElements.define(`media-volume-range`,hu);var gu=hu;function _u(e){return`
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
    `}function vu(){return E(`Loop`)}var yu=class extends G{constructor(){super(...arguments),this.container=null}static get observedAttributes(){return[...super.observedAttributes,b.MEDIA_LOOP]}connectedCallback(){super.connectedCallback(),this.container=this.shadowRoot?.querySelector(`#icon`)||null,this.container&&(this.container.textContent=E(`Loop`))}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===b.MEDIA_LOOP&&this.container&&this.setAttribute(`aria-checked`,this.mediaLoop?`true`:`false`)}get mediaLoop(){return M(this,b.MEDIA_LOOP)}set mediaLoop(e){N(this,b.MEDIA_LOOP,e)}handleClick(){let e=!this.mediaLoop,t=new D.CustomEvent(v.MEDIA_LOOP_REQUEST,{composed:!0,bubbles:!0,detail:e});this.dispatchEvent(t)}};yu.getSlotTemplateHTML=_u,yu.getTooltipContentHTML=vu,D.customElements.get(`media-loop-button`)||D.customElements.define(`media-loop-button`,yu);var bu=yu;function Q(e){if(typeof e==`boolean`)return e?``:void 0;if(typeof e!=`function`){if(Array.isArray(e)&&e.every(e=>typeof e==`string`||typeof e==`number`||typeof e==`boolean`))return e.join(` `);if(!(typeof e==`object`&&e))return e}}_({tagName:`media-gesture-receiver`,elementClass:Nt,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-container`,elementClass:fn,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var xu=_({tagName:`media-controller`,elementClass:Zr,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});_({tagName:`media-tooltip`,elementClass:ti,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-chrome-button`,elementClass:_i,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-airplay-button`,elementClass:Ci,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var Su=_({tagName:`media-captions-button`,elementClass:Mi,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});_({tagName:`media-cast-button`,elementClass:zi,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-chrome-dialog`,elementClass:ua,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-chrome-range`,elementClass:Ha,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var Cu=_({tagName:`media-control-bar`,elementClass:Xa,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});_({tagName:`media-text-display`,elementClass:ao,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-duration-display`,elementClass:mo,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-error-dialog`,elementClass:Oo,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-keyboard-shortcuts-dialog`,elementClass:Lo,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var wu=_({tagName:`media-fullscreen-button`,elementClass:Yo,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});_({tagName:`media-live-button`,elementClass:is,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-loading-indicator`,elementClass:gs,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var Tu=_({tagName:`media-mute-button`,elementClass:ws,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Eu=_({tagName:`media-pip-button`,elementClass:As,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Du=_({tagName:`media-playback-rate-button`,elementClass:Vs,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Ou=_({tagName:`media-play-button`,elementClass:Js,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});_({tagName:`media-poster-image`,elementClass:ec,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-preview-chapter-display`,elementClass:sc,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-preview-thumbnail`,elementClass:hc,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),_({tagName:`media-preview-time-display`,elementClass:Sc,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var ku=_({tagName:`media-seek-backward-button`,elementClass:jc,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Au=_({tagName:`media-seek-forward-button`,elementClass:Bc,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),ju=_({tagName:`media-time-display`,elementClass:fl,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Mu=_({tagName:`media-time-range`,elementClass:su,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Nu=_({tagName:`media-volume-range`,elementClass:gu,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});_({tagName:`media-loop-button`,elementClass:bu,react:g.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var $=o(),Pu={noTooltip:!0};function Fu({isMobile:e,showPip:t,showCaptionsButton:n,topChrome:r,extraControls:i,children:a,className:o,style:s}){return(0,$.jsxs)(xu,{className:`ts-player ${o||``}`.trim(),autohide:`2`,style:s,children:[a,(0,$.jsx)(`div`,{slot:`top-chrome`,className:`ts-player-top`,children:r}),(0,$.jsxs)(Cu,{className:`ts-player-bar`,children:[(0,$.jsx)(Ou,{...Pu,className:e?`ts-player-play-lg`:void 0}),(0,$.jsx)(ku,{...Pu,seekOffset:10}),(0,$.jsx)(Au,{...Pu,seekOffset:10}),(0,$.jsx)(ju,{showDuration:!0}),(0,$.jsx)(Mu,{}),(0,$.jsx)(Tu,{...Pu}),e?null:(0,$.jsx)(Nu,{}),i,n?(0,$.jsx)(Su,{...Pu}):null,(0,$.jsx)(Du,{...Pu,rates:[.75,1,1.25,1.5,2]}),t?(0,$.jsx)(Eu,{...Pu}):null,(0,$.jsx)(wu,{...Pu})]})]})}var Iu=e=>!!(e.canPlayType(`application/vnd.apple.mpegurl`)||e.canPlayType(`application/x-mpegURL`));function Lu({enabled:e,open:t,videoRef:n,videoEpoch:r,src:i,playbackRate:a=1,onLoading:o,onError:s,onNotSupported:c,onSubtitleTracks:l,onSubtitleTrack:u,onDuration:d}){let m=(0,g.useRef)(null),h=(0,g.useRef)(o),ee=(0,g.useRef)(s),te=(0,g.useRef)(c),ne=(0,g.useRef)(l),re=(0,g.useRef)(u),ie=(0,g.useRef)(d);return(0,g.useEffect)(()=>{h.current=o,ee.current=s,te.current=c,ne.current=l,re.current=u,ie.current=d}),(0,g.useEffect)(()=>{let r=n.current;if(!e||!t||!r||!i)return;let o=null,s=!1;h.current?.(!0),ne.current?.([]);let c=()=>{let e=(o?.subtitleTracks||[]).map((e,t)=>({id:t,name:e.name,lang:e.lang}));ne.current?.(e)};if(Te.isSupported()){let e=f();o=new Te({xhrSetup:t=>{e&&t.setRequestHeader(`Authorization`,e)}}),m.current=o,o.on(Te.Events.MANIFEST_PARSED,()=>{c(),re.current?.(o?.subtitleTrack??-1),r.playbackRate=a,r.play().catch(()=>{})}),o.on(Te.Events.LEVEL_LOADED,(e,t)=>{let n=t.details?.totalduration;typeof n==`number`&&ie.current?.(n)}),o.on(Te.Events.SUBTITLE_TRACKS_UPDATED,c),o.on(Te.Events.SUBTITLE_TRACK_SWITCH,(e,t)=>re.current?.(t.id)),o.on(Te.Events.ERROR,(e,t)=>{t.fatal&&o&&(t.type===Te.ErrorTypes.NETWORK_ERROR?o.startLoad():t.type===Te.ErrorTypes.MEDIA_ERROR?o.recoverMediaError():(o.stopLoad(),h.current?.(!1),ee.current?.()))}),o.loadSource(i),o.attachMedia(r)}else Iu(r)?(s=!0,r.src=p(i),r.load(),r.playbackRate=a,r.play().catch(()=>{})):te.current?.();return()=>{m.current===o&&(m.current=null),o?.destroy(),s&&(r.pause(),r.removeAttribute(`src`),r.load()),ne.current?.([]),re.current?.(-1)}},[e,t,n,r,i]),m}var Ru=5e3,zu=5;function Bu({enabled:e,hash:t,fileIndex:n,title:r,initialTimecode:i=0,videoRef:a,onViewedChange:o}){let s=(0,g.useRef)(0),c=(0,g.useRef)(!1),l=(0,g.useRef)(o);(0,g.useEffect)(()=>{l.current=o},[o]),(0,g.useEffect)(()=>{c.current=!1},[t,n,i]);let u=(0,g.useCallback)(async i=>{if(e&&t&&n!=null)try{await ye(t,n,i),xe({hash:t,fileIndex:n,title:r||t,fileName:r||String(n),timecode:i}),l.current?.()}catch{}},[e,t,n,r]);return{flushTimecode:(0,g.useCallback)(()=>{let t=a.current;t&&e&&u(t.currentTime)},[a,e,u]),onTimeUpdate:(0,g.useCallback)(()=>{let t=a.current;if(!t||!e)return;let n=Date.now();n-s.current<Ru||(s.current=n,u(t.currentTime))},[a,e,u]),applyResumeIfNeeded:(0,g.useCallback)(()=>{let e=a.current;if(!e||c.current)return;if(!(i>zu)){c.current=!0;return}let t=e.duration;if(Number.isFinite(t)&&t>0&&i>=t-zu){c.current=!0;return}e.currentTime=i,c.current=!0},[a,i]),saveTimecode:u}}var Vu=3e4;function Hu({tip:e,children:t}){return(0,$.jsx)(`span`,{className:`inline-flex min-w-0 max-w-full items-center gap-1`,title:e,children:t})}function Uu(e){switch(e.split(`?`)[0].split(`.`).pop()?.toLowerCase()){case`mp4`:return`video/mp4`;case`ogg`:case`ogv`:return`video/ogg`;case`webm`:return`video/webm`;default:return``}}var Wu=e=>!!(e.canPlayType(`application/vnd.apple.mpegurl`)||e.canPlayType(`application/x-mpegURL`)),Gu=()=>typeof document<`u`&&`pictureInPictureEnabled`in document&&!!document.pictureInPictureEnabled;function Ku({videoSrc:e,downloadSrc:n=e,title:o,onNotSupported:f,hls:h=!1,heartbeatSrc:ye=``,showTrigger:xe=!0,inlineTrigger:Ee=!1,inlineTriggerPrimary:Oe=!1,initiallyOpen:ke=!1,onClose:Ae,captionSrc:je,hash:_,fileIndex:Me,initialTimecode:v=0,trackTimecode:y=!1,onViewedChange:Ne,audioTracks:Pe=[],audioIndex:b=0,onAudioIndexChange:Fe}){let{t:x}=ne(),S=be(),{setImmersive:Ie}=u(),Le=(0,g.useRef)(null),C=(0,g.useRef)(f),w=(0,g.useRef)(null),[T,Re]=(0,g.useState)(ke),[ze,Be]=(0,g.useState)(0),[Ve,He]=(0,g.useState)(!0),[Ue,We]=(0,g.useState)(!1),[Ge,Ke]=(0,g.useState)(!1),[qe,Je]=(0,g.useState)(!1),[E,Ye]=(0,g.useState)([]),[Xe,Ze]=(0,g.useState)(!1),[Qe,$e]=(0,g.useState)(b),[et,tt]=(0,g.useState)(e),[nt,rt]=(0,g.useState)(Pe),{data:it}=we(T?_:void 0),at=Ce(_,{enabled:T&&!!_,fast:!1}),ot=!!(_&&Me!=null&&y),D=h&&!!_&&Me!=null&&nt.length>1,O=Gu(),st=h?E.length>0:!!je,ct=S?pe:qe?he:ue,lt=S?void 0:qe?`min(86dvh, calc(100dvh - 4rem))`:`min(72dvh, 40rem)`,ut=(0,g.useCallback)(e=>{Le.current=e,Be(e=>e+1)},[]);(0,g.useEffect)(()=>{C.current=f},[f]),(0,g.useEffect)(()=>{tt(e),$e(b),rt(Pe),We(!1),He(!0)},[e,b,Pe]),Lu({enabled:h,open:T,videoRef:Le,videoEpoch:ze,src:et,onLoading:He,onError:()=>We(!0),onNotSupported:()=>C.current?.(),onSubtitleTracks:Ye});let{flushTimecode:dt,onTimeUpdate:ft,applyResumeIfNeeded:pt}=Bu({enabled:ot,hash:_,fileIndex:Me,title:o,initialTimecode:v,videoRef:Le,onViewedChange:Ne}),mt=(0,g.useCallback)(()=>Re(!0),[]),ht=(0,g.useCallback)(()=>{dt(),Re(!1),We(!1),Ze(!1),Je(!1),Ae?.()},[dt,Ae]),gt=c({isOpen:T,onOpenChange:e=>{e?Re(!0):ht()}});d(T),(0,g.useEffect)(()=>{if(!S||!T){Ie(!1);return}return Ie(!0),()=>Ie(!1)},[S,T,Ie]),(0,g.useEffect)(()=>{let e=document.createElement(`video`);(h?Te.isSupported()||Wu(e):e.canPlayType(Uu(et)))||C.current?.()},[h,et]),(0,g.useEffect)(()=>{let e=Le.current;if(T&&!h&&e)return He(!0),e.src=p(et),e.load(),e.play().catch(()=>{}),()=>{e.pause(),e.removeAttribute(`src`),e.load()}},[T,h,ze,et]),(0,g.useEffect)(()=>{if(!T||!ye)return;let e=window.setInterval(()=>{m(ye,{cache:`no-store`}).catch(()=>{})},Vu);return()=>window.clearInterval(e)},[ye,T]),(0,g.useEffect)(()=>{let e=()=>Ke(!!document.fullscreenElement);return document.addEventListener(`fullscreenchange`,e),()=>document.removeEventListener(`fullscreenchange`,e)},[]),(0,g.useEffect)(()=>{let e=Le.current;if(!e||w.current==null)return;let t=w.current,n=()=>{w.current!=null&&(e.currentTime=t,w.current=null,e.play().catch(()=>{}))};return e.addEventListener(`loadedmetadata`,n,{once:!0}),()=>e.removeEventListener(`loadedmetadata`,n)},[ze,et]),(0,g.useEffect)(()=>{if(!T||!h||!_||Me==null)return;if(Pe.length){rt(Pe);return}let e=!1;return l.get(fe(_,Me)).then(({data:t})=>{e||rt(ge(t))}).catch(()=>{}),()=>{e=!0}},[T,h,_,Me,Pe]);let _t=e=>{if(!_||Me==null)return;let t=Le.current;t&&(w.current=t.currentTime),$e(e),tt(Se(_,Me,e)),Ze(!1),Fe?.(e)},vt=se,yt=it?.download_speed??0,bt=it?.active_peers,xt=it?.total_peers??0,St=it?.connected_seeders??0,Ct=bt==null?`—`:`${bt}/${xt}`,k=bt==null?`—`:`↑${St}`,wt=ae(at.Filled,at.Capacity)??`—`,Tt=_==null?null:(0,$.jsxs)(`div`,{className:`mt-0.5 flex min-w-0 flex-wrap items-center gap-x-2.5 gap-y-1 text-[11px] leading-none text-white/80 drop-shadow sm:text-xs`,"aria-label":`${x(`DownloadSpeed`)}, ${x(`Peers`)}, ${x(`CacheFilled`)}`,children:[(0,$.jsxs)(Hu,{tip:x(`DownloadSpeed`),children:[(0,$.jsx)(ve,{className:`size-3 shrink-0`,strokeWidth:2.25,"aria-hidden":!0}),(0,$.jsx)(`span`,{className:`tabular-nums`,children:oe(yt)})]}),(0,$.jsxs)(Hu,{tip:x(`Peers`),children:[(0,$.jsx)(_e,{className:`size-3 shrink-0`,strokeWidth:2,"aria-hidden":!0}),(0,$.jsx)(`span`,{className:`tabular-nums`,children:Ct}),(0,$.jsx)(`span`,{className:`tabular-nums text-white/70`,children:k})]}),(0,$.jsxs)(Hu,{tip:x(`CacheHint`),children:[(0,$.jsx)(ee,{className:`size-3 shrink-0`,strokeWidth:2,"aria-hidden":!0}),(0,$.jsx)(`span`,{className:`min-w-0 truncate tabular-nums`,children:wt})]})]}),A=D?(0,$.jsxs)(a,{isOpen:Xe,onOpenChange:Ze,children:[(0,$.jsx)(a.Trigger,{children:(0,$.jsx)(t,{isIconOnly:!0,variant:`ghost`,className:`${me} text-white hover-fine:bg-white/15`,"aria-label":x(`SelectAudioTrack`),children:(0,$.jsx)(le,{...vt,"aria-hidden":!0})})}),(0,$.jsxs)(a.Content,{placement:`bottom end`,className:`max-h-72 w-72 overflow-y-auto border border-white/10 bg-neutral-950/95 p-1 backdrop-blur-xl`,children:[(0,$.jsx)(`p`,{className:`px-2 py-1.5 text-[11px] font-semibold tracking-wide text-muted uppercase`,children:x(`SelectAudioTrack`)}),nt.map((e,n)=>{let{title:r,meta:i}=de(e,n);return(0,$.jsx)(t,{variant:n===Qe?`secondary`:`ghost`,className:`h-auto w-full justify-start gap-2 py-2`,onPress:()=>_t(n),children:(0,$.jsxs)(`span`,{className:`min-w-0 flex-1 text-left`,children:[(0,$.jsx)(`span`,{className:`block truncate text-sm font-medium`,children:r}),i?(0,$.jsx)(`span`,{className:`mt-0.5 block truncate text-xs text-muted`,children:i}):null]})},n)})]})]}):null,j=S?null:(0,$.jsx)(t,{isIconOnly:!0,variant:`ghost`,className:`${me} text-white hover-fine:bg-white/15`,"aria-label":x(qe?`ExitFullscreen`:`ExpandPlayer`),onPress:()=>Je(e=>!e),children:qe?(0,$.jsx)(De,{...vt,"aria-hidden":!0}):(0,$.jsx)(te,{...vt,"aria-hidden":!0})});return(0,$.jsxs)($.Fragment,{children:[xe&&(Ee?(0,$.jsxs)(t,{variant:Oe?`primary`:`secondary`,size:`sm`,onPress:mt,className:Oe?`min-h-10 shrink-0 px-3`:`min-h-10 min-w-[72px] max-w-full flex-1`,children:[Oe?(0,$.jsx)(ce,{...ie,fill:`currentColor`,"aria-hidden":!0}):null,x(`Play`)]}):(0,$.jsxs)(t,{variant:`secondary`,onPress:mt,children:[(0,$.jsx)(ce,{...ie,"aria-hidden":!0}),x(`Play`)]})),(0,$.jsx)(r,{state:gt,children:(0,$.jsx)(r.Backdrop,{isDismissable:!Ge,className:`bg-black/75 backdrop-blur-sm`,children:(0,$.jsx)(r.Container,{size:S?`full`:`lg`,scroll:`inside`,placement:S?`center`:`auto`,className:S?`ts-player-modal-full p-0`:`p-4 sm:p-5`,children:(0,$.jsx)(r.Dialog,{className:S?`overflow-hidden rounded-none border-0 bg-black text-white shadow-none`:`overflow-hidden border border-white/10 bg-[#0a0e0c] text-white shadow-2xl shadow-black/50`,style:ct,children:(0,$.jsxs)(r.Body,{className:S?`flex h-full min-h-0 flex-col gap-0 p-0`:`gap-0 p-0`,children:[Ue?(0,$.jsxs)(`div`,{className:`space-y-3 p-4`,children:[(0,$.jsx)(s,{status:`danger`,children:x(`PlaybackError`)}),(0,$.jsx)(t,{variant:`secondary`,onPress:()=>window.open(n,`_blank`,`noopener,noreferrer`),children:x(`OpenLink`)})]}):(0,$.jsx)(Fu,{isMobile:S,showPip:O,showCaptionsButton:st,className:S?`ts-player--mobile`:`ts-player--desktop`,style:S?{width:`100%`,height:`100%`}:{width:`100%`,maxHeight:lt,aspectRatio:`16 / 9`},topChrome:(0,$.jsxs)(`div`,{className:`flex items-start gap-2 px-3 pb-10 pt-[max(0.75rem,env(safe-area-inset-top))] pl-[max(0.75rem,env(safe-area-inset-left))] pr-[max(0.75rem,env(safe-area-inset-right))]`,children:[(0,$.jsxs)(`div`,{className:`min-w-0 flex-1`,children:[(0,$.jsx)(`p`,{className:`truncate text-sm font-semibold text-white drop-shadow`,title:o||x(`Play`),children:o||x(`Play`)}),Tt]}),A,j,(0,$.jsx)(t,{isIconOnly:!0,variant:`ghost`,className:`${me} text-white hover-fine:bg-white/15`,"aria-label":x(`Close`),onPress:ht,children:(0,$.jsx)(re,{...vt,"aria-hidden":!0})})]}),extraControls:null,children:(0,$.jsx)(`video`,{slot:`media`,ref:ut,autoPlay:!0,playsInline:!0,crossOrigin:`anonymous`,className:`ts-player-video`,onTimeUpdate:()=>{ft(),Ve&&He(!1)},onLoadedMetadata:()=>{He(!1),We(!1),pt()},onWaiting:()=>He(!0),onCanPlay:()=>{He(!1),We(!1)},onPlaying:()=>He(!1),onPause:dt,onError:()=>{He(!1),We(!0)},children:!h&&je?(0,$.jsx)(`track`,{kind:`captions`,src:je,srcLang:`und`,label:`Captions`,default:!0}):null},et)}),Ve&&!Ue?(0,$.jsx)(`div`,{className:`pointer-events-none absolute inset-0 z-10 grid place-items-center bg-black/55`,children:(0,$.jsxs)(`div`,{className:`flex flex-col items-center gap-3 px-4 text-center`,children:[(0,$.jsx)(i,{size:`lg`,color:`current`,className:`text-accent`}),(0,$.jsx)(`p`,{className:`text-sm font-medium text-white/90`,children:x(`Buffering`)})]})}):null]})})})})})]})}export{Ku as default};