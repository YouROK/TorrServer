import{i as e}from"./rolldown-runtime-BgaNhQyE.js";import{C as t,N as n,O as r,b as i,j as a,l as o,o as s,r as c,x as l}from"./heroui-FW202rcf.js";import{i as u}from"./vendor-DX_-Mk4S.js";import{n as d,r as f}from"./ModalOpenContext-8MfYxuwx.js";import{a as p,l as m,n as h}from"./authCredentials-bzDUNYl6.js";import{t as g}from"./createLucideIcon-CFKXZzfw.js";import{n as ee,t as te}from"./maximize-2-CqPjvNae.js";import{t as ne}from"./useTranslation-DyfSBSYg.js";import{a as re,t as ie}from"./format-DVo6TP4W.js";import{G as ae,H as oe,V as se,W as ce,a as le,c as ue,i as de,it as fe,mt as pe,s as me,tt as he,v as ge,vt as _e,w as ve,y as ye}from"./index-UQPIK195.js";import{d as be,l as xe,u as Se}from"./dialogSizes-DYo2rf0H.js";import{r as Ce,t as we}from"./useTorrentDetail-CZ-8O_nI.js";import{t as Te}from"./hls-BgLaj4xL.js";var Ee=g(`minimize-2`,[[`path`,{d:`m14 10 7-7`,key:`oa77jy`}],[`path`,{d:`M20 10h-6V4`,key:`mjg0md`}],[`path`,{d:`m3 21 7-7`,key:`tjx5ai`}],[`path`,{d:`M4 14h6v6`,key:`rmj7iw`}]]),_=e(n(),1),De=`torrserver:now-playing`,Oe=e=>{try{window.dispatchEvent(new CustomEvent(De,{detail:e}))}catch{}},ke=new Set([`style`,`children`,`ref`,`key`,`suppressContentEditableWarning`,`suppressHydrationWarning`,`dangerouslySetInnerHTML`]),Ae={className:`class`,htmlFor:`for`};function je(e){return e.toLowerCase()}function Me(e){if(typeof e==`boolean`)return e?``:void 0;if(typeof e!=`function`&&!(typeof e==`object`&&e))return e}function v({react:e,tagName:t,elementClass:n,events:r,displayName:i,defaultProps:a,toAttributeName:o=je,toAttributeValue:s=Me}){let c=Number.parseInt(e.version)>=19,l=e.forwardRef((i,l)=>{let u=e.useRef(null),d=e.useRef(new Map),f={},p={},m={},h={};for(let[e,t]of Object.entries(i)){if(ke.has(e)){m[e]=t;continue}let r=o(Ae[e]??e);if(n.prototype&&e in n.prototype&&!(e in(globalThis.HTMLElement?.prototype??{}))&&!n.observedAttributes?.some(e=>e===r)){h[e]=t;continue}if(e.startsWith(`on`)){f[e]=t;continue}let i=s(t);r&&i!=null&&(p[r]=String(i),c||(m[r]=i)),r&&c&&(i===Me(t)?m[r]=t:m[r]=i)}if(typeof window<`u`){for(let t in f){let n=f[t],i=t.endsWith(`Capture`),a=(r?.[t]??t.slice(2).toLowerCase()).slice(0,i?-7:void 0);e.useLayoutEffect(()=>{let e=u?.current;if(!(!e||typeof n!=`function`))return e.addEventListener(a,n,i),()=>{e.removeEventListener(a,n,i)}},[u?.current,n])}e.useLayoutEffect(()=>{if(u.current===null)return;let e=new Map;for(let t in h)Ne(u.current,t,h[t]),d.current.delete(t),e.set(t,h[t]);for(let[e,t]of d.current)Ne(u.current,e,void 0);d.current=e})}if(typeof window>`u`&&n?.getTemplateHTML&&n?.shadowRootOptions){let{mode:t,delegatesFocus:r}=n.shadowRootOptions;m.children=[e.createElement(`template`,{shadowrootmode:t,shadowrootdelegatesfocus:r,dangerouslySetInnerHTML:{__html:n.getTemplateHTML(p,i)},key:`ce-la-react-ssr-template-shadow-root`}),m.children]}return e.createElement(t,{...a,...m,ref:e.useCallback(e=>{u.current=e,typeof l==`function`?l(e):l!==null&&(l.current=e)},[l])},m.children)});return l.displayName=i??n.name,l}function Ne(e,t,n){e[t]=n,n==null&&t in(globalThis.HTMLElement?.prototype??{})&&e.removeAttribute(t)}var y={MEDIA_PLAY_REQUEST:`mediaplayrequest`,MEDIA_PAUSE_REQUEST:`mediapauserequest`,MEDIA_MUTE_REQUEST:`mediamuterequest`,MEDIA_UNMUTE_REQUEST:`mediaunmuterequest`,MEDIA_LOOP_REQUEST:`medialooprequest`,MEDIA_VOLUME_REQUEST:`mediavolumerequest`,MEDIA_SEEK_REQUEST:`mediaseekrequest`,MEDIA_AIRPLAY_REQUEST:`mediaairplayrequest`,MEDIA_ENTER_FULLSCREEN_REQUEST:`mediaenterfullscreenrequest`,MEDIA_EXIT_FULLSCREEN_REQUEST:`mediaexitfullscreenrequest`,MEDIA_PREVIEW_REQUEST:`mediapreviewrequest`,MEDIA_ENTER_PIP_REQUEST:`mediaenterpiprequest`,MEDIA_EXIT_PIP_REQUEST:`mediaexitpiprequest`,MEDIA_ENTER_CAST_REQUEST:`mediaentercastrequest`,MEDIA_EXIT_CAST_REQUEST:`mediaexitcastrequest`,MEDIA_SHOW_TEXT_TRACKS_REQUEST:`mediashowtexttracksrequest`,MEDIA_HIDE_TEXT_TRACKS_REQUEST:`mediahidetexttracksrequest`,MEDIA_SHOW_SUBTITLES_REQUEST:`mediashowsubtitlesrequest`,MEDIA_DISABLE_SUBTITLES_REQUEST:`mediadisablesubtitlesrequest`,MEDIA_TOGGLE_SUBTITLES_REQUEST:`mediatogglesubtitlesrequest`,MEDIA_PLAYBACK_RATE_REQUEST:`mediaplaybackraterequest`,MEDIA_RENDITION_REQUEST:`mediarenditionrequest`,MEDIA_AUDIO_TRACK_REQUEST:`mediaaudiotrackrequest`,MEDIA_SEEK_TO_LIVE_REQUEST:`mediaseektoliverequest`,REGISTER_MEDIA_STATE_RECEIVER:`registermediastatereceiver`,UNREGISTER_MEDIA_STATE_RECEIVER:`unregistermediastatereceiver`},b={MEDIA_CHROME_ATTRIBUTES:`mediachromeattributes`,MEDIA_CONTROLLER:`mediacontroller`},Pe={MEDIA_AIRPLAY_UNAVAILABLE:`mediaAirplayUnavailable`,MEDIA_AUDIO_TRACK_ENABLED:`mediaAudioTrackEnabled`,MEDIA_AUDIO_TRACK_LIST:`mediaAudioTrackList`,MEDIA_AUDIO_TRACK_UNAVAILABLE:`mediaAudioTrackUnavailable`,MEDIA_BUFFERED:`mediaBuffered`,MEDIA_CAST_UNAVAILABLE:`mediaCastUnavailable`,MEDIA_CHAPTERS_CUES:`mediaChaptersCues`,MEDIA_CURRENT_TIME:`mediaCurrentTime`,MEDIA_DURATION:`mediaDuration`,MEDIA_ENDED:`mediaEnded`,MEDIA_ERROR:`mediaError`,MEDIA_ERROR_CODE:`mediaErrorCode`,MEDIA_ERROR_MESSAGE:`mediaErrorMessage`,MEDIA_FULLSCREEN_UNAVAILABLE:`mediaFullscreenUnavailable`,MEDIA_HAS_PLAYED:`mediaHasPlayed`,MEDIA_HEIGHT:`mediaHeight`,MEDIA_IS_AIRPLAYING:`mediaIsAirplaying`,MEDIA_IS_CASTING:`mediaIsCasting`,MEDIA_IS_FULLSCREEN:`mediaIsFullscreen`,MEDIA_IS_PIP:`mediaIsPip`,MEDIA_LOADING:`mediaLoading`,MEDIA_MUTED:`mediaMuted`,MEDIA_LOOP:`mediaLoop`,MEDIA_PAUSED:`mediaPaused`,MEDIA_PIP_UNAVAILABLE:`mediaPipUnavailable`,MEDIA_PLAYBACK_RATE:`mediaPlaybackRate`,MEDIA_PREVIEW_CHAPTER:`mediaPreviewChapter`,MEDIA_PREVIEW_COORDS:`mediaPreviewCoords`,MEDIA_PREVIEW_IMAGE:`mediaPreviewImage`,MEDIA_PREVIEW_TIME:`mediaPreviewTime`,MEDIA_RENDITION_LIST:`mediaRenditionList`,MEDIA_RENDITION_SELECTED:`mediaRenditionSelected`,MEDIA_RENDITION_UNAVAILABLE:`mediaRenditionUnavailable`,MEDIA_SEEKABLE:`mediaSeekable`,MEDIA_STREAM_TYPE:`mediaStreamType`,MEDIA_SUBTITLES_LIST:`mediaSubtitlesList`,MEDIA_SUBTITLES_SHOWING:`mediaSubtitlesShowing`,MEDIA_TARGET_LIVE_WINDOW:`mediaTargetLiveWindow`,MEDIA_TIME_IS_LIVE:`mediaTimeIsLive`,MEDIA_VOLUME:`mediaVolume`,MEDIA_VOLUME_LEVEL:`mediaVolumeLevel`,MEDIA_VOLUME_UNAVAILABLE:`mediaVolumeUnavailable`,MEDIA_LANG:`mediaLang`,MEDIA_WIDTH:`mediaWidth`},Fe=Object.entries(Pe),x=Fe.reduce((e,[t,n])=>(e[t]=n.toLowerCase(),e),{}),Ie=Fe.reduce((e,[t,n])=>(e[t]=n.toLowerCase(),e),{USER_INACTIVE_CHANGE:`userinactivechange`,BREAKPOINTS_CHANGE:`breakpointchange`,BREAKPOINTS_COMPUTED:`breakpointscomputed`});Object.entries(Ie).reduce((e,[t,n])=>{let r=x[t];return r&&(e[n]=r),e},{userinactivechange:`userinactive`});var S=Object.entries(x).reduce((e,[t,n])=>{let r=Ie[t];return r&&(e[n]=r),e},{userinactive:`userinactivechange`}),C={SUBTITLES:`subtitles`,CAPTIONS:`captions`,DESCRIPTIONS:`descriptions`,CHAPTERS:`chapters`,METADATA:`metadata`},Le={DISABLED:`disabled`,HIDDEN:`hidden`,SHOWING:`showing`},Re={MOUSE:`mouse`,PEN:`pen`,TOUCH:`touch`},w={UNAVAILABLE:`unavailable`,UNSUPPORTED:`unsupported`},T={LIVE:`live`,ON_DEMAND:`on-demand`,UNKNOWN:`unknown`},E={INLINE:`inline`,FULLSCREEN:`fullscreen`,PICTURE_IN_PICTURE:`picture-in-picture`};function ze(e){return e?.map(Be).join(` `)}function Be(e){if(e){let{id:t,width:n,height:r}=e;return[t,n,r].filter(e=>e!=null).join(`:`)}}function Ve(e){return e?.map(He).join(` `)}function He(e){if(e){let{id:t,kind:n,language:r,label:i}=e;return[t,n,r,i].filter(e=>e!=null).join(`:`)}}function Ue(e){return typeof e==`number`&&!Number.isNaN(e)&&Number.isFinite(e)}var We=e=>new Promise(t=>setTimeout(t,e)),Ge={en:{"Start airplay":`Start airplay`,"Stop airplay":`Stop airplay`,Audio:`Audio`,Captions:`Captions`,"Enable captions":`Enable captions`,"Disable captions":`Disable captions`,"Start casting":`Start casting`,"Stop casting":`Stop casting`,"Enter fullscreen mode":`Enter fullscreen mode`,"Exit fullscreen mode":`Exit fullscreen mode`,Mute:`Mute`,Unmute:`Unmute`,Loop:`Loop`,"Enter picture in picture mode":`Enter picture in picture mode`,"Exit picture in picture mode":`Exit picture in picture mode`,Play:`Play`,Pause:`Pause`,"Playback rate":`Playback rate`,"Playback rate {playbackRate}":`Playback rate {playbackRate}`,Quality:`Quality`,"Seek backward":`Seek backward`,"Seek forward":`Seek forward`,Settings:`Settings`,Auto:`Auto`,"audio player":`audio player`,"video player":`video player`,volume:`volume`,seek:`seek`,"closed captions":`closed captions`,"current playback rate":`current playback rate`,"playback time":`playback time`,"media loading":`media loading`,settings:`settings`,"audio tracks":`audio tracks`,quality:`quality`,play:`play`,pause:`pause`,mute:`mute`,unmute:`unmute`,"chapter: {chapterName}":`chapter: {chapterName}`,live:`live`,Off:`Off`,"start airplay":`start airplay`,"stop airplay":`stop airplay`,"start casting":`start casting`,"stop casting":`stop casting`,"enter fullscreen mode":`enter fullscreen mode`,"exit fullscreen mode":`exit fullscreen mode`,"enter picture in picture mode":`enter picture in picture mode`,"exit picture in picture mode":`exit picture in picture mode`,"seek to live":`seek to live`,"playing live":`playing live`,"seek back {seekOffset} seconds":`seek back {seekOffset} seconds`,"seek forward {seekOffset} seconds":`seek forward {seekOffset} seconds`,"Network Error":`Network Error`,"Decode Error":`Decode Error`,"Source Not Supported":`Source Not Supported`,"Encryption Error":`Encryption Error`,"A network error caused the media download to fail.":`A network error caused the media download to fail.`,"A media error caused playback to be aborted. The media could be corrupt or your browser does not support this format.":`A media error caused playback to be aborted. The media could be corrupt or your browser does not support this format.`,"An unsupported error occurred. The server or network failed, or your browser does not support this format.":`An unsupported error occurred. The server or network failed, or your browser does not support this format.`,"The media is encrypted and there are no keys to decrypt it.":`The media is encrypted and there are no keys to decrypt it.`,hour:`hour`,hours:`hours`,minute:`minute`,minutes:`minutes`,second:`second`,seconds:`seconds`,"{time} remaining":`{time} remaining`,"{currentTime} of {totalTime}":`{currentTime} of {totalTime}`,"video not loaded, unknown time.":`video not loaded, unknown time.`}},Ke=globalThis.navigator?.language||`en`,qe=e=>{Ke=e},Je=e=>{let[t]=Ke.split(`-`);return Ge[Ke]?.[e]||Ge[t]?.[e]||Ge.en?.[e]||e},Ye=()=>{let[e]=Ke.split(`-`);return Ge[Ke]?Ke:Ge[e]?e:`en`},D=(e,t={})=>Je(e).replace(/\{(\w+)\}/g,(e,n)=>n in t?String(t[n]):`{${n}}`),Xe=[{singular:`hour`,plural:`hours`},{singular:`minute`,plural:`minutes`},{singular:`second`,plural:`seconds`}],Ze=(e,t)=>`${e} ${D(e===1?Xe[t].singular:Xe[t].plural)}`,Qe=e=>{if(!Ue(e))return``;let t=Math.abs(e),n=t!==e,r=new Date(0,0,0,0,0,t,0),i=[r.getHours(),r.getMinutes(),r.getSeconds()].map((e,t)=>e&&Ze(e,t)).filter(e=>e).join(`, `);return n?D(`{time} remaining`,{time:i}):i};function $e(e,t){let n=!1;e<0&&(n=!0,e=0-e),e=e<0?0:e;let r=Math.floor(e%60),i=Math.floor(e/60%60),a=Math.floor(e/3600),o=Math.floor(t/60%60),s=Math.floor(t/3600);return(isNaN(e)||e===1/0)&&(a=i=r=`0`),a=a>0||s>0?a+`:`:``,i=((a||o>=10)&&i<10?`0`+i:i)+`:`,r=r<10?`0`+r:r,(n?`-`:``)+a+i+r}Object.freeze({length:0,start(e){let t=e>>>0;if(t>=this.length)throw new DOMException(`Failed to execute 'start' on 'TimeRanges': The index provided (${t}) is greater than or equal to the maximum bound (${this.length}).`);return 0},end(e){let t=e>>>0;if(t>=this.length)throw new DOMException(`Failed to execute 'end' on 'TimeRanges': The index provided (${t}) is greater than or equal to the maximum bound (${this.length}).`);return 0}});var et=class{addEventListener(){}removeEventListener(){}dispatchEvent(){return!0}},tt=class extends et{},nt=class extends tt{constructor(){super(...arguments),this.role=null}},rt=class{observe(){}unobserve(){}disconnect(){}},it={createElement:function(){return new at.HTMLElement},createElementNS:function(){return new at.HTMLElement},addEventListener(){},removeEventListener(){},dispatchEvent(e){return!1}},at={ResizeObserver:rt,document:it,Node:tt,Element:nt,HTMLElement:class extends nt{constructor(){super(...arguments),this.innerHTML=``}get content(){return new at.DocumentFragment}},DocumentFragment:class extends et{},customElements:{get:function(){},define:function(){},whenDefined:function(){}},localStorage:{getItem(e){return null},setItem(e,t){},removeItem(e){}},CustomEvent:function(){},getComputedStyle:function(){},navigator:{languages:[],get userAgent(){return``}},matchMedia(e){return{matches:!1,media:e}},DOMParser:class{parseFromString(e,t){return{body:{textContent:e}}}}},ot=`global`in globalThis&&(globalThis==null?void 0:globalThis.global)===globalThis||typeof window>`u`||window.customElements===void 0,st=Object.keys(at).every(e=>e in globalThis),O=ot&&!st?at:globalThis,k=ot&&!st?it:globalThis.document,ct=new WeakMap,lt=e=>{let t=ct.get(e);return t||ct.set(e,t=new Set),t},ut=new O.ResizeObserver(e=>{for(let t of e)for(let e of lt(t.target))e(t)});function dt(e,t){lt(e).add(t),ut.observe(e)}function ft(e,t){let n=lt(e);n.delete(t),n.size||ut.unobserve(e)}function pt(e){let t={};for(let n of e)t[n.name]=n.value;return t}function mt(e){return ht(e)??bt(e,`media-controller`)}function ht(e){let{MEDIA_CONTROLLER:t}=b,n=e.getAttribute(t);if(n)return St(e)?.getElementById(n)}var gt=(e,t,n=`.value`)=>{let r=e.querySelector(n);r&&(r.textContent=t)},_t=(e,t)=>{let n=`slot[name="${t}"]`,r=e.shadowRoot.querySelector(n);return r?r.children:[]},vt=(e,t)=>_t(e,t)[0],yt=(e,t)=>!e||!t?!1:e?.contains(t)?!0:yt(e,t.getRootNode().host),bt=(e,t)=>e?e.closest(t)||bt(e.getRootNode().host,t):null;function xt(e=document){let t=e?.activeElement;return t?xt(t.shadowRoot)??t:null}function St(e){let t=(e?.getRootNode)?.call(e);return t instanceof ShadowRoot||t instanceof Document?t:null}function Ct(e,{depth:t=3,checkOpacity:n=!0,checkVisibilityCSS:r=!0}={}){if(e.checkVisibility)return e.checkVisibility({checkOpacity:n,checkVisibilityCSS:r});let i=e;for(;i&&t>0;){let e=getComputedStyle(i);if(n&&e.opacity===`0`||r&&e.visibility===`hidden`||e.display===`none`)return!1;i=i.parentElement,t--}return!0}function wt(e,t,n,r){let i=r.x-n.x,a=r.y-n.y,o=i*i+a*a;if(o===0)return 0;let s=((e-n.x)*i+(t-n.y)*a)/o;return Math.max(0,Math.min(1,s))}function A(e,t){return Tt(e,e=>e===t)||Et(e,t)}function Tt(e,t){let n;for(n of e.querySelectorAll(`style:not([media])`)??[]){let e;try{e=n.sheet?.cssRules}catch{continue}for(let n of e??[])if(t(n.selectorText))return n}}function Et(e,t){let n=e.querySelectorAll(`style:not([media])`)??[],r=n?.[n.length-1];if(!r?.sheet)return console.warn(`Media Chrome: No style sheet found on style tag of`,e),{style:{setProperty:()=>{},removeProperty:()=>``,getPropertyValue:()=>``}};let i=r?.sheet.insertRule(`${t}{}`,r.sheet.cssRules.length);return r.sheet.cssRules?.[i]}function j(e,t,n=NaN){let r=e.getAttribute(t);return r==null?n:+r}function M(e,t,n){let r=+n;if(n==null||Number.isNaN(r)){e.hasAttribute(t)&&e.removeAttribute(t);return}j(e,t,void 0)!==r&&e.setAttribute(t,`${r}`)}function N(e,t){return e.hasAttribute(t)}function P(e,t,n){if(n==null){e.hasAttribute(t)&&e.removeAttribute(t);return}N(e,t)!=n&&e.toggleAttribute(t,n)}function F(e,t,n=null){return e.getAttribute(t)??n}function I(e,t,n){if(n==null){e.hasAttribute(t)&&e.removeAttribute(t);return}let r=`${n}`;F(e,t,void 0)!==r&&e.setAttribute(t,r)}var Dt=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Ot=(e,t,n)=>(Dt(e,t,`read from private field`),n?n.call(e):t.get(e)),kt=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},At=(e,t,n,r)=>(Dt(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),L;function jt(e){return`
    <style>
      :host {
        display: var(--media-control-display, var(--media-gesture-receiver-display, inline-block));
        box-sizing: border-box;
      }
    </style>
  `}var Mt=class extends O.HTMLElement{constructor(){if(super(),kt(this,L,void 0),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}}static get observedAttributes(){return[b.MEDIA_CONTROLLER,x.MEDIA_PAUSED]}attributeChangedCallback(e,t,n){var r,i,a,o;e===b.MEDIA_CONTROLLER&&(t&&((i=(r=Ot(this,L))?.unassociateElement)==null||i.call(r,this),At(this,L,null)),n&&this.isConnected&&(At(this,L,this.getRootNode()?.getElementById(n)),(o=(a=Ot(this,L))?.associateElement)==null||o.call(a,this)))}connectedCallback(){var e,t;this.tabIndex=-1,this.setAttribute(`aria-hidden`,`true`),At(this,L,Nt(this)),this.getAttribute(b.MEDIA_CONTROLLER)&&((t=(e=Ot(this,L))?.associateElement)==null||t.call(e,this)),Ot(this,L)&&(Ot(this,L).addEventListener(`pointerdown`,this),Ot(this,L).addEventListener(`click`,this),Ot(this,L).hasAttribute(`tabindex`)||(Ot(this,L).tabIndex=0))}disconnectedCallback(){var e,t,n,r;this.getAttribute(b.MEDIA_CONTROLLER)&&((t=(e=Ot(this,L))?.unassociateElement)==null||t.call(e,this)),(n=Ot(this,L))==null||n.removeEventListener(`pointerdown`,this),(r=Ot(this,L))==null||r.removeEventListener(`click`,this),At(this,L,null)}handleEvent(e){let t=e.composedPath()?.[0];if([`video`,`media-controller`].includes(t?.localName)){if(e.type===`pointerdown`)this._pointerType=e.pointerType;else if(e.type===`click`){let{clientX:t,clientY:n}=e,{left:r,top:i,width:a,height:o}=this.getBoundingClientRect(),s=t-r,c=n-i;if(s<0||c<0||s>a||c>o||a===0&&o===0)return;let l=this._pointerType||`mouse`;if(this._pointerType=void 0,l===Re.TOUCH){this.handleTap(e);return}else if(l===Re.MOUSE||l===Re.PEN){this.handleMouseClick(e);return}}}}get mediaPaused(){return N(this,x.MEDIA_PAUSED)}set mediaPaused(e){P(this,x.MEDIA_PAUSED,e)}handleTap(e){}handleMouseClick(e){let t=this.mediaPaused?y.MEDIA_PLAY_REQUEST:y.MEDIA_PAUSE_REQUEST;this.dispatchEvent(new O.CustomEvent(t,{composed:!0,bubbles:!0}))}};L=new WeakMap,Mt.shadowRootOptions={mode:`open`},Mt.getTemplateHTML=jt;function Nt(e){let t=e.getAttribute(b.MEDIA_CONTROLLER);return t?e.getRootNode()?.getElementById(t):bt(e,`media-controller`)}O.customElements.get(`media-gesture-receiver`)||O.customElements.define(`media-gesture-receiver`,Mt);var Pt=Mt,Ft=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},R=(e,t,n)=>(Ft(e,t,`read from private field`),n?n.call(e):t.get(e)),z=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},It=(e,t,n,r)=>(Ft(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),Lt=(e,t,n)=>(Ft(e,t,`access private method`),n),Rt,zt,Bt,Vt,Ht,Ut,Wt,Gt,Kt,qt,Jt,Yt,Xt,Zt,Qt,$t,en,tn,nn,rn,B={AUDIO:`audio`,AUTOHIDE:`autohide`,BREAKPOINTS:`breakpoints`,GESTURES_DISABLED:`gesturesdisabled`,KEYBOARD_CONTROL:`keyboardcontrol`,NO_AUTOHIDE:`noautohide`,USER_INACTIVE:`userinactive`,AUTOHIDE_OVER_CONTROLS:`autohideovercontrols`};function an(e){return`
    <style>
      
      :host([${x.MEDIA_IS_FULLSCREEN}]) ::slotted([slot=media]) {
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

      
      :host([${B.USER_INACTIVE}]:not([${x.MEDIA_PAUSED}]):not([${x.MEDIA_IS_AIRPLAYING}]):not([${x.MEDIA_IS_CASTING}]):not([${B.AUDIO}])) ::slotted(:not([slot=media]):not([slot=poster]):not([${B.NO_AUTOHIDE}]):not([role=dialog])) {
        opacity: 0;
        transition: var(--media-control-transition-out, opacity 1s);
      }

      :host([${B.USER_INACTIVE}]:not([${B.NO_AUTOHIDE}]):not([${x.MEDIA_PAUSED}]):not([${x.MEDIA_IS_CASTING}]):not([${B.AUDIO}])) ::slotted([slot=media]) {
        cursor: none;
      }

      :host([${B.USER_INACTIVE}][${B.AUTOHIDE_OVER_CONTROLS}]:not([${B.NO_AUTOHIDE}]):not([${x.MEDIA_PAUSED}]):not([${x.MEDIA_IS_CASTING}]):not([${B.AUDIO}])) * {
        --media-cursor: none;
        cursor: none;
      }


      ::slotted(media-control-bar)  {
        align-self: stretch;
      }

      
      :host(:not([${B.AUDIO}])[${x.MEDIA_HAS_PLAYED}]) slot[name=poster] {
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
  `}var on=Object.values(x),sn=`sm:384 md:576 lg:768 xl:960`;function cn(e){ln(e.target,e.contentRect.width)}function ln(e,t){if(!e.isConnected)return;let n=un(e.getAttribute(B.BREAKPOINTS)??sn),r=dn(n,t),i=!1;if(Object.keys(n).forEach(t=>{if(r.includes(t)){e.hasAttribute(`breakpoint${t}`)||(e.setAttribute(`breakpoint${t}`,``),i=!0);return}e.hasAttribute(`breakpoint${t}`)&&(e.removeAttribute(`breakpoint${t}`),i=!0)}),i){let t=new CustomEvent(Ie.BREAKPOINTS_CHANGE,{detail:r});e.dispatchEvent(t)}e.breakpointsComputed||(e.breakpointsComputed=!0,e.dispatchEvent(new CustomEvent(Ie.BREAKPOINTS_COMPUTED,{bubbles:!0,composed:!0})))}function un(e){let t=e.split(/\s+/);return Object.fromEntries(t.map(e=>e.split(`:`)))}function dn(e,t){return Object.keys(e).filter(n=>t>=parseInt(e[n]))}var fn=class extends O.HTMLElement{constructor(){if(super(),z(this,Kt),z(this,Jt),z(this,Xt),z(this,Qt),z(this,en),z(this,Rt,void 0),z(this,zt,0),z(this,Bt,null),z(this,Vt,null),z(this,Ht,void 0),this.breakpointsComputed=!1,z(this,Ut,e=>{let t=this.media;for(let n of e){if(n.type!==`childList`)continue;let e=n.removedNodes;for(let r of e){if(r.slot!=`media`||n.target!=this)continue;let e=n.previousSibling&&n.previousSibling.previousElementSibling;if(!e||!t)this.mediaUnsetCallback(r);else{let t=e.slot!==`media`;for(;(e=e.previousSibling)!==null;)e.slot==`media`&&(t=!1);t&&this.mediaUnsetCallback(r)}}if(t)for(let e of n.addedNodes)e===t&&this.handleMediaUpdated(t)}}),z(this,Wt,!1),z(this,Gt,e=>{R(this,Wt)||(setTimeout(()=>{cn(e),It(this,Wt,!1)},0),It(this,Wt,!0))}),z(this,nn,void 0),z(this,rn,()=>{if(!R(this,nn).assignedElements({flatten:!0}).length){R(this,Bt)&&this.mediaUnsetCallback(R(this,Bt));return}this.handleMediaUpdated(this.media)}),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes),t=this.constructor.getTemplateHTML(e);this.shadowRoot.setHTMLUnsafe?this.shadowRoot.setHTMLUnsafe(t):this.shadowRoot.innerHTML=t}It(this,Rt,new MutationObserver(R(this,Ut)))}static get observedAttributes(){return[B.AUTOHIDE,B.GESTURES_DISABLED].concat(on).filter(e=>![x.MEDIA_RENDITION_LIST,x.MEDIA_AUDIO_TRACK_LIST,x.MEDIA_CHAPTERS_CUES,x.MEDIA_WIDTH,x.MEDIA_HEIGHT,x.MEDIA_ERROR,x.MEDIA_ERROR_MESSAGE].includes(e))}attributeChangedCallback(e,t,n){e.toLowerCase()==B.AUTOHIDE&&(this.autohide=n)}get media(){let e=this.querySelector(`:scope > [slot=media]`);return e?.nodeName==`SLOT`&&(e=e.assignedElements({flatten:!0})[0]),e}async handleMediaUpdated(e){e&&(It(this,Bt,e),e.localName.includes(`-`)&&await O.customElements.whenDefined(e.localName),this.mediaSetCallback(e))}connectedCallback(){var e;R(this,Rt).observe(this,{childList:!0,subtree:!0}),dt(this,R(this,Gt));let t=this.getAttribute(B.AUDIO)==null?D(`video player`):D(`audio player`);this.setAttribute(`role`,`region`),this.setAttribute(`aria-label`,t),this.handleMediaUpdated(this.media),this.setAttribute(B.USER_INACTIVE,``),ln(this,this.getBoundingClientRect().width);let n=this.querySelector(`:scope > slot[slot=media]`);n&&(It(this,nn,n),R(this,nn).addEventListener(`slotchange`,R(this,rn))),this.addEventListener(`pointerdown`,this),this.addEventListener(`pointermove`,this),this.addEventListener(`pointerup`,this),this.addEventListener(`mouseleave`,this),this.addEventListener(`keyup`,this),(e=O.window)==null||e.addEventListener(`mouseup`,this)}disconnectedCallback(){var e;ft(this,R(this,Gt)),clearTimeout(R(this,Vt)),R(this,Rt).disconnect(),this.media&&this.mediaUnsetCallback(this.media),(e=O.window)==null||e.removeEventListener(`mouseup`,this),this.removeEventListener(`pointerdown`,this),this.removeEventListener(`pointermove`,this),this.removeEventListener(`pointerup`,this),this.removeEventListener(`mouseleave`,this),this.removeEventListener(`keyup`,this),R(this,nn)&&(R(this,nn).removeEventListener(`slotchange`,R(this,rn)),It(this,nn,null)),It(this,Wt,!1)}mediaSetCallback(e){}mediaUnsetCallback(e){It(this,Bt,null)}handleEvent(e){switch(e.type){case`pointerdown`:It(this,zt,e.timeStamp);break;case`pointermove`:Lt(this,Kt,qt).call(this,e);break;case`pointerup`:Lt(this,Jt,Yt).call(this,e);break;case`mouseleave`:Lt(this,Xt,Zt).call(this);break;case`mouseup`:this.removeAttribute(B.KEYBOARD_CONTROL);break;case`keyup`:Lt(this,en,tn).call(this),this.setAttribute(B.KEYBOARD_CONTROL,``);break}}set autohide(e){let t=Number(e);It(this,Ht,isNaN(t)?0:t)}get autohide(){return(R(this,Ht)===void 0?2:R(this,Ht)).toString()}get breakpoints(){return F(this,B.BREAKPOINTS)}set breakpoints(e){I(this,B.BREAKPOINTS,e)}get audio(){return N(this,B.AUDIO)}set audio(e){P(this,B.AUDIO,e)}get gesturesDisabled(){return N(this,B.GESTURES_DISABLED)}set gesturesDisabled(e){P(this,B.GESTURES_DISABLED,e)}get keyboardControl(){return N(this,B.KEYBOARD_CONTROL)}set keyboardControl(e){P(this,B.KEYBOARD_CONTROL,e)}get noAutohide(){return N(this,B.NO_AUTOHIDE)}set noAutohide(e){P(this,B.NO_AUTOHIDE,e)}get autohideOverControls(){return N(this,B.AUTOHIDE_OVER_CONTROLS)}set autohideOverControls(e){P(this,B.AUTOHIDE_OVER_CONTROLS,e)}get userInteractive(){return N(this,B.USER_INACTIVE)}set userInteractive(e){P(this,B.USER_INACTIVE,e)}};Rt=new WeakMap,zt=new WeakMap,Bt=new WeakMap,Vt=new WeakMap,Ht=new WeakMap,Ut=new WeakMap,Wt=new WeakMap,Gt=new WeakMap,Kt=new WeakSet,qt=function(e){if(e.pointerType!==`mouse`&&e.timeStamp-R(this,zt)<250)return;Lt(this,Qt,$t).call(this),clearTimeout(R(this,Vt));let t=this.hasAttribute(B.AUTOHIDE_OVER_CONTROLS);([this,this.media].includes(e.target)||t)&&Lt(this,en,tn).call(this)},Jt=new WeakSet,Yt=function(e){if(e.pointerType===`touch`){let t=!this.hasAttribute(B.USER_INACTIVE);[this,this.media].includes(e.target)&&t?Lt(this,Xt,Zt).call(this):Lt(this,en,tn).call(this)}else e.composedPath().some(e=>[`media-play-button`,`media-fullscreen-button`].includes(e?.localName))&&Lt(this,en,tn).call(this)},Xt=new WeakSet,Zt=function(){if(R(this,Ht)<0||this.hasAttribute(B.USER_INACTIVE))return;this.setAttribute(B.USER_INACTIVE,``);let e=new O.CustomEvent(Ie.USER_INACTIVE_CHANGE,{composed:!0,bubbles:!0,detail:!0});this.dispatchEvent(e)},Qt=new WeakSet,$t=function(){if(!this.hasAttribute(B.USER_INACTIVE))return;this.removeAttribute(B.USER_INACTIVE);let e=new O.CustomEvent(Ie.USER_INACTIVE_CHANGE,{composed:!0,bubbles:!0,detail:!1});this.dispatchEvent(e)},en=new WeakSet,tn=function(){Lt(this,Qt,$t).call(this),clearTimeout(R(this,Vt));let e=parseInt(this.autohide);e<0||It(this,Vt,setTimeout(()=>{Lt(this,Xt,Zt).call(this)},e*1e3))},nn=new WeakMap,rn=new WeakMap,fn.shadowRootOptions={mode:`open`},fn.getTemplateHTML=an,O.customElements.get(`media-container`)||O.customElements.define(`media-container`,fn);var pn=fn,mn=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},V=(e,t,n)=>(mn(e,t,`read from private field`),n?n.call(e):t.get(e)),hn=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},gn=(e,t,n,r)=>(mn(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),_n,vn,yn,bn,xn,Sn,Cn=class{constructor(e,t,{defaultValue:n}={defaultValue:void 0}){hn(this,xn),hn(this,_n,void 0),hn(this,vn,void 0),hn(this,yn,void 0),hn(this,bn,new Set),gn(this,_n,e),gn(this,vn,t),gn(this,yn,new Set(n))}[Symbol.iterator](){return V(this,xn,Sn).values()}get length(){return V(this,xn,Sn).size}get value(){return[...V(this,xn,Sn)].join(` `)??``}set value(e){e!==this.value&&(gn(this,bn,new Set),this.add(...e?.split(` `)??[]))}toString(){return this.value}item(e){return[...V(this,xn,Sn)][e]}values(){return V(this,xn,Sn).values()}forEach(e,t){V(this,xn,Sn).forEach(e,t)}add(...e){var t;e.forEach(e=>V(this,bn).add(e)),!(this.value===``&&!V(this,_n)?.hasAttribute(`${V(this,vn)}`))&&((t=V(this,_n))==null||t.setAttribute(`${V(this,vn)}`,`${this.value}`))}remove(...e){var t;e.forEach(e=>V(this,bn).delete(e)),(t=V(this,_n))==null||t.setAttribute(`${V(this,vn)}`,`${this.value}`)}contains(e){return V(this,xn,Sn).has(e)}toggle(e,t){return t===void 0?this.contains(e)?(this.remove(e),!1):(this.add(e),!0):t?(this.add(e),!0):(this.remove(e),!1)}replace(e,t){return this.remove(e),this.add(t),e===t}};_n=new WeakMap,vn=new WeakMap,yn=new WeakMap,bn=new WeakMap,xn=new WeakSet,Sn=function(){return V(this,bn).size?V(this,bn):V(this,yn)};var wn=(e=``)=>e.split(/\s+/),Tn=(e=``)=>{let[t,n,r]=e.split(`:`),i=r?decodeURIComponent(r):void 0;return{kind:t===`cc`?C.CAPTIONS:C.SUBTITLES,language:n,label:i}},En=(e=``,t={})=>wn(e).map(e=>{let n=Tn(e);return{...t,...n}}),Dn=e=>e?Array.isArray(e)?e.map(e=>typeof e==`string`?Tn(e):e):typeof e==`string`?En(e):[e]:[],On=({kind:e,label:t,language:n}={kind:`subtitles`})=>t?`${e===`captions`?`cc`:`sb`}:${n}:${encodeURIComponent(t)}`:n,kn=(e=[])=>Array.prototype.map.call(e,On).join(` `),An=(e,t)=>n=>n[e]===t,jn=e=>{let t=Object.entries(e).map(([e,t])=>An(e,t));return e=>t.every(t=>t(e))},Mn=(e,t=[],n=[])=>{let r=Dn(n).map(jn);Array.from(t).filter(e=>r.some(t=>t(e))).forEach(t=>{t.mode=e})},Nn=(e,t=()=>!0)=>{if(!e?.textTracks)return[];let n=typeof t==`function`?t:jn(t);return Array.from(e.textTracks).filter(n)},Pn=e=>!!e.mediaSubtitlesShowing?.length||e.hasAttribute(x.MEDIA_SUBTITLES_SHOWING),Fn=e=>{let{media:t,fullscreenElement:n}=e;try{let e=n&&`requestFullscreen`in n?`requestFullscreen`:n&&`webkitRequestFullScreen`in n?`webkitRequestFullScreen`:void 0;if(e){let t=n[e]?.call(n);if(t instanceof Promise)return t.catch(()=>{})}else t?.webkitEnterFullscreen?t.webkitEnterFullscreen():t?.requestFullscreen&&t.requestFullscreen()}catch(e){console.error(e)}},In=`exitFullscreen`in k?`exitFullscreen`:`webkitExitFullscreen`in k?`webkitExitFullscreen`:`webkitCancelFullScreen`in k?`webkitCancelFullScreen`:void 0,Ln=e=>{let{documentElement:t}=e;if(In){let e=(t?.[In])?.call(t);if(e instanceof Promise)return e.catch(()=>{})}},Rn=`fullscreenElement`in k?`fullscreenElement`:`webkitFullscreenElement`in k?`webkitFullscreenElement`:void 0,zn=e=>{let{documentElement:t,media:n}=e,r=t?.[Rn];return!r&&`webkitDisplayingFullscreen`in n&&`webkitPresentationMode`in n&&n.webkitDisplayingFullscreen&&n.webkitPresentationMode===E.FULLSCREEN?n:r},Bn=e=>{let{media:t,documentElement:n,fullscreenElement:r=t}=e;if(!t||!n)return!1;let i=zn(e);if(!i)return!1;if(i===r||i===t)return!0;if(i.localName.includes(`-`)){let e=i.shadowRoot;if(!(Rn in e))return yt(i,r);for(;e?.[Rn];){if(e[Rn]===r)return!0;e=e[Rn]?.shadowRoot}}return!1},Vn=`fullscreenEnabled`in k?`fullscreenEnabled`:`webkitFullscreenEnabled`in k?`webkitFullscreenEnabled`:void 0,Hn=e=>{let{documentElement:t,media:n}=e;return!!t?.[Vn]||n&&`webkitSupportsFullscreen`in n},Un,Wn=()=>{var e;return Un||(Un=((e=k)?.createElement)?.call(e,`video`),Un)},Gn=async(e=Wn())=>{if(!e)return!1;let t=e.volume;e.volume=t/2+.1;let n=new AbortController,r=await Promise.race([Kn(e,n.signal),qn(e,t)]);return n.abort(),r},Kn=(e,t)=>new Promise(n=>{e.addEventListener(`volumechange`,()=>n(!0),{signal:t})}),qn=async(e,t)=>{for(let n=0;n<10;n++){if(e.volume===t)return!1;await We(10)}return e.volume!==t},Jn=/.*Version\/.*Safari\/.*/.test(O.navigator.userAgent),Yn=(e=Wn())=>O.matchMedia(`(display-mode: standalone)`).matches&&Jn?!1:typeof e?.requestPictureInPicture==`function`,Xn=(e=Wn())=>Hn({documentElement:k,media:e}),Zn=Xn(),Qn=Yn(),$n=!!O.WebKitPlaybackTargetAvailabilityEvent,er=!!O.chrome,tr=e=>Nn(e.media,e=>[C.SUBTITLES,C.CAPTIONS].includes(e.kind)).sort((e,t)=>e.kind>=t.kind?1:-1),nr=e=>Nn(e.media,e=>e.mode===Le.SHOWING&&[C.SUBTITLES,C.CAPTIONS].includes(e.kind)),rr=(e,t)=>{let n=tr(e),r=nr(e),i=!!r.length;if(n.length){if(t===!1||i&&t!==!0)Mn(Le.DISABLED,n,r);else if(t===!0||!i&&t!==!1){let t=n[0],{options:i}=e;if(!i?.noSubtitlesLangPref){let e=O.localStorage.getItem(`media-chrome-pref-subtitles-lang`),r=e?[e,...O.navigator.languages]:O.navigator.languages,i=n.filter(e=>r.some(t=>e.language.toLowerCase().startsWith(t.split(`-`)[0]))).sort((e,t)=>r.findIndex(t=>e.language.toLowerCase().startsWith(t.split(`-`)[0]))-r.findIndex(e=>t.language.toLowerCase().startsWith(e.split(`-`)[0])));i[0]&&(t=i[0])}let{language:a,label:o,kind:s}=t;Mn(Le.DISABLED,n,r),Mn(Le.SHOWING,n,[{language:a,label:o,kind:s}])}}},ir=(e,t)=>e===t?!0:e==null||t==null||typeof e!=typeof t?!1:typeof e==`number`&&Number.isNaN(e)&&Number.isNaN(t)?!0:typeof e==`object`?Array.isArray(e)?ar(e,t):Object.entries(e).every(([e,n])=>e in t&&ir(n,t[e])):!1,ar=(e,t)=>{let n=Array.isArray(e),r=Array.isArray(t);return n===r?n||r?e.length===t.length&&e.every((e,n)=>ir(e,t[n])):!0:!1},or=Object.values(T),sr,cr=Gn().then(e=>(sr=e,sr)),lr=async(...e)=>{await Promise.all(e.filter(e=>e).map(async e=>{if(!(`localName`in e&&e instanceof O.HTMLElement))return;let t=e.localName;if(!t.includes(`-`))return;let n=O.customElements.get(t);n&&e instanceof n||(await O.customElements.whenDefined(t),O.customElements.upgrade(e))}))},ur=new O.DOMParser,dr=e=>e&&(ur.parseFromString(e,`text/html`).body.textContent||e),fr={mediaError:{get(e,t){let{media:n}=e;if(t?.type!==`playing`)return n?.error},mediaEvents:[`emptied`,`error`,`playing`]},mediaErrorCode:{get(e,t){let{media:n}=e;if(t?.type!==`playing`)return n?.error?.code},mediaEvents:[`emptied`,`error`,`playing`]},mediaErrorMessage:{get(e,t){let{media:n}=e;if(t?.type!==`playing`)return n?.error?.message??``},mediaEvents:[`emptied`,`error`,`playing`]},mediaWidth:{get(e){let{media:t}=e;return t?.videoWidth??0},mediaEvents:[`resize`]},mediaHeight:{get(e){let{media:t}=e;return t?.videoHeight??0},mediaEvents:[`resize`]},mediaPaused:{get(e){let{media:t}=e;return t?.paused??!0},set(e,t){var n;let{media:r}=t;r&&(e?r.pause():(n=r.play())==null||n.catch(()=>{}))},mediaEvents:[`play`,`playing`,`pause`,`emptied`]},mediaHasPlayed:{get(e,t){let{media:n}=e;return n?t?t.type===`playing`:!n.paused:!1},mediaEvents:[`playing`,`emptied`]},mediaEnded:{get(e){let{media:t}=e;return t?.ended??!1},mediaEvents:[`seeked`,`ended`,`emptied`]},mediaPlaybackRate:{get(e){let{media:t}=e;return t?.playbackRate??1},set(e,t){let{media:n}=t;n&&Number.isFinite(+e)&&(n.playbackRate=+e)},mediaEvents:[`ratechange`,`loadstart`]},mediaMuted:{get(e){let{media:t}=e;return t?.muted??!1},set(e,t){let{media:n,options:{noMutedPref:r}={}}=t;if(n){n.muted=e;try{let t=O.localStorage.getItem(`media-chrome-pref-muted`)!==null,i=n.hasAttribute(`muted`);if(r){t&&O.localStorage.removeItem(`media-chrome-pref-muted`);return}if(i&&!t)return;O.localStorage.setItem(`media-chrome-pref-muted`,e?`true`:`false`)}catch(e){console.debug(`Error setting muted pref`,e)}}},mediaEvents:[`volumechange`],stateOwnersUpdateHandlers:[(e,t)=>{let{options:{noMutedPref:n}}=t,{media:r}=t;if(!(!r||r.muted||n))try{let n=O.localStorage.getItem(`media-chrome-pref-muted`)===`true`;fr.mediaMuted.set(n,t),e(n)}catch(e){console.debug(`Error getting muted pref`,e)}}]},mediaLoop:{get(e){let{media:t}=e;return t?.loop},set(e,t){let{media:n}=t;n&&(n.loop=e)},mediaEvents:[`medialooprequest`]},mediaVolume:{get(e){let{media:t}=e;return t?.volume??1},set(e,t){let{media:n,options:{noVolumePref:r}={}}=t;if(n){try{e==null?O.localStorage.removeItem(`media-chrome-pref-volume`):!n.hasAttribute(`muted`)&&!r&&O.localStorage.setItem(`media-chrome-pref-volume`,e.toString())}catch(e){console.debug(`Error setting volume pref`,e)}Number.isFinite(+e)&&(n.volume=+e)}},mediaEvents:[`volumechange`],stateOwnersUpdateHandlers:[(e,t)=>{let{options:{noVolumePref:n}}=t;if(!n)try{let{media:n}=t;if(!n)return;let r=O.localStorage.getItem(`media-chrome-pref-volume`);if(r==null)return;fr.mediaVolume.set(+r,t),e(+r)}catch(e){console.debug(`Error getting volume pref`,e)}}]},mediaVolumeLevel:{get(e){let{media:t}=e;return t?.volume===void 0?`high`:t.muted||t.volume===0?`off`:t.volume<.5?`low`:t.volume<.75?`medium`:`high`},mediaEvents:[`volumechange`]},mediaCurrentTime:{get(e){let{media:t}=e;return t?.currentTime??0},set(e,t){let{media:n}=t;!n||!Ue(e)||(n.currentTime=e)},mediaEvents:[`timeupdate`,`loadedmetadata`]},mediaDuration:{get(e){let{media:t,options:{defaultDuration:n}={}}=e;return n&&(!t||!t.duration||Number.isNaN(t.duration)||!Number.isFinite(t.duration))?n:Number.isFinite(t?.duration)?t.duration:NaN},mediaEvents:[`durationchange`,`loadedmetadata`,`emptied`]},mediaLoading:{get(e){let{media:t}=e;return t?.readyState<3},mediaEvents:[`waiting`,`playing`,`emptied`]},mediaSeekable:{get(e){let{media:t}=e;if(!t?.seekable?.length)return;let n=t.seekable.start(0),r=t.seekable.end(t.seekable.length-1);if(!(!n&&!r))return[Number(n.toFixed(3)),Number(r.toFixed(3))]},mediaEvents:[`loadedmetadata`,`emptied`,`progress`,`seekablechange`]},mediaBuffered:{get(e){let{media:t}=e,n=t?.buffered??[];return Array.from(n).map((e,t)=>[Number(n.start(t).toFixed(3)),Number(n.end(t).toFixed(3))])},mediaEvents:[`progress`,`emptied`]},mediaStreamType:{get(e){let{media:t,options:{defaultStreamType:n}={}}=e,r=[T.LIVE,T.ON_DEMAND].includes(n)?n:void 0;if(!t)return r;let{streamType:i}=t;if(or.includes(i))return i===T.UNKNOWN?r:i;let a=t.duration;return a===1/0?T.LIVE:Number.isFinite(a)?T.ON_DEMAND:r},mediaEvents:[`emptied`,`durationchange`,`loadedmetadata`,`streamtypechange`]},mediaTargetLiveWindow:{get(e){let{media:t}=e;if(!t)return NaN;let{targetLiveWindow:n}=t,r=fr.mediaStreamType.get(e);return(n==null||Number.isNaN(n))&&r===T.LIVE?0:n},mediaEvents:[`emptied`,`durationchange`,`loadedmetadata`,`streamtypechange`,`targetlivewindowchange`]},mediaTimeIsLive:{get(e){let{media:t,options:{liveEdgeOffset:n=10}={}}=e;if(!t)return!1;if(typeof t.liveEdgeStart==`number`)return!Number.isNaN(t.liveEdgeStart)&&t.currentTime>=t.liveEdgeStart;if(fr.mediaStreamType.get(e)!==T.LIVE)return!1;let r=t.seekable;if(!r)return!0;if(!r.length)return!1;let i=r.end(r.length-1)-n;return t.currentTime>=i},mediaEvents:[`playing`,`timeupdate`,`progress`,`waiting`,`emptied`]},mediaSubtitlesList:{get(e){return tr(e).map(({kind:e,label:t,language:n})=>({kind:e,label:t,language:n}))},mediaEvents:[`loadstart`],textTracksEvents:[`addtrack`,`removetrack`]},mediaSubtitlesShowing:{get(e){return nr(e).map(({kind:e,label:t,language:n})=>({kind:e,label:t,language:n}))},mediaEvents:[`loadstart`],textTracksEvents:[`addtrack`,`removetrack`,`change`],stateOwnersUpdateHandlers:[(e,t)=>{var n,r;let{media:i,options:a}=t;if(!i)return;let o=e=>{a.defaultSubtitles&&(e&&![C.CAPTIONS,C.SUBTITLES].includes(e?.track?.kind)||rr(t,!0))};return i.addEventListener(`loadstart`,o),(n=i.textTracks)==null||n.addEventListener(`addtrack`,o),(r=i.textTracks)==null||r.addEventListener(`removetrack`,o),()=>{var e,t;i.removeEventListener(`loadstart`,o),(e=i.textTracks)==null||e.removeEventListener(`addtrack`,o),(t=i.textTracks)==null||t.removeEventListener(`removetrack`,o)}}]},mediaChaptersCues:{get(e){let{media:t}=e;if(!t)return[];let[n]=Nn(t,{kind:C.CHAPTERS});return Array.from(n?.cues??[]).map(({text:e,startTime:t,endTime:n})=>({text:dr(e),startTime:t,endTime:n}))},mediaEvents:[`loadstart`,`loadedmetadata`],textTracksEvents:[`addtrack`,`removetrack`,`change`],stateOwnersUpdateHandlers:[(e,t)=>{let{media:n}=t;if(!n)return;let r=n.querySelector(`track[kind="chapters"][default][src]`),i=n.shadowRoot?.querySelector(`:is(video,audio) > track[kind="chapters"][default][src]`);return r?.addEventListener(`load`,e),i?.addEventListener(`load`,e),()=>{r?.removeEventListener(`load`,e),i?.removeEventListener(`load`,e)}}]},mediaIsPip:{get(e){let{media:t,documentElement:n}=e;if(!t||!n||!n.pictureInPictureElement)return!1;if(n.pictureInPictureElement===t)return!0;if(n.pictureInPictureElement instanceof HTMLMediaElement)return t.localName?.includes(`-`)?yt(t,n.pictureInPictureElement):!1;if(n.pictureInPictureElement.localName.includes(`-`)){let e=n.pictureInPictureElement.shadowRoot;for(;e?.pictureInPictureElement;){if(e.pictureInPictureElement===t)return!0;e=e.pictureInPictureElement?.shadowRoot}}return!1},set(e,t){let{media:n}=t;if(n)if(e){if(!k.pictureInPictureEnabled){console.warn(`MediaChrome: Picture-in-picture is not enabled`);return}if(!n.requestPictureInPicture){console.warn(`MediaChrome: The current media does not support picture-in-picture`);return}let e=()=>{console.warn(`MediaChrome: The media is not ready for picture-in-picture. It must have a readyState > 0.`)};n.requestPictureInPicture().catch(t=>{if(t.code===11){if(!n.src){console.warn(`MediaChrome: The media is not ready for picture-in-picture. It must have a src set.`);return}if(n.readyState===0&&n.preload===`none`){let t=()=>{n.removeEventListener(`loadedmetadata`,r),n.preload=`none`},r=()=>{n.requestPictureInPicture().catch(e),t()};n.addEventListener(`loadedmetadata`,r),n.preload=`metadata`,setTimeout(()=>{n.readyState===0&&e(),t()},1e3)}else throw t}else throw t})}else k.pictureInPictureElement&&k.exitPictureInPicture()},mediaEvents:[`enterpictureinpicture`,`leavepictureinpicture`]},mediaRenditionList:{get(e){let{media:t}=e;return[...t?.videoRenditions??[]].map(e=>({...e}))},mediaEvents:[`emptied`,`loadstart`],videoRenditionsEvents:[`addrendition`,`removerendition`]},mediaRenditionSelected:{get(e){let{media:t}=e;return t?.videoRenditions?.[t.videoRenditions?.selectedIndex]?.id},set(e,t){let{media:n}=t;if(!n?.videoRenditions){console.warn(`MediaController: Rendition selection not supported by this media.`);return}let r=e,i=Array.prototype.findIndex.call(n.videoRenditions,e=>e.id==r);n.videoRenditions.selectedIndex!=i&&(n.videoRenditions.selectedIndex=i)},mediaEvents:[`emptied`],videoRenditionsEvents:[`addrendition`,`removerendition`,`change`]},mediaAudioTrackList:{get(e){let{media:t}=e;return[...t?.audioTracks??[]]},mediaEvents:[`emptied`,`loadstart`],audioTracksEvents:[`addtrack`,`removetrack`]},mediaAudioTrackEnabled:{get(e){let{media:t}=e;return[...t?.audioTracks??[]].find(e=>e.enabled)?.id},set(e,t){let{media:n}=t;if(!n?.audioTracks){console.warn(`MediaChrome: Audio track selection not supported by this media.`);return}let r=e;for(let e of n.audioTracks)e.enabled=r==e.id},mediaEvents:[`emptied`],audioTracksEvents:[`addtrack`,`removetrack`,`change`]},mediaIsFullscreen:{get(e){return Bn(e)},set(e,t,n){var r;e?(Fn(t),n.detail&&!t.media?.inert&&((r=t.media)==null||r.focus())):Ln(t)},rootEvents:[`fullscreenchange`,`webkitfullscreenchange`],mediaEvents:[`webkitbeginfullscreen`,`webkitendfullscreen`,`webkitpresentationmodechanged`]},mediaIsCasting:{get(e){let{media:t}=e;return!t?.remote||t.remote?.state===`disconnected`?!1:t.remote.state===`connected`},set(e,t){let{media:n}=t;if(n&&!(e&&n.remote?.state!==`disconnected`)&&!(!e&&n.remote?.state!==`connected`)){if(typeof n.remote.prompt!=`function`){console.warn(`MediaChrome: Casting is not supported in this environment`);return}n.remote.prompt().catch(()=>{})}},remoteEvents:[`connect`,`connecting`,`disconnect`]},mediaIsAirplaying:{get(){return!1},set(e,t){let{media:n}=t;if(n){if(!(n.webkitShowPlaybackTargetPicker&&O.WebKitPlaybackTargetAvailabilityEvent)){console.error(`MediaChrome: received a request to select AirPlay but AirPlay is not supported in this environment`);return}n.webkitShowPlaybackTargetPicker()}},mediaEvents:[`webkitcurrentplaybacktargetiswirelesschanged`]},mediaFullscreenUnavailable:{get(e){let{media:t}=e;if(!Zn||!Xn(t))return w.UNSUPPORTED}},mediaPipUnavailable:{get(e){let{media:t}=e;if(!Qn||!Yn(t))return w.UNSUPPORTED;if(t?.disablePictureInPicture)return w.UNAVAILABLE}},mediaVolumeUnavailable:{get(e){let{media:t}=e;if(sr===!1||t?.volume==null)return w.UNSUPPORTED},stateOwnersUpdateHandlers:[e=>{sr??cr.then(t=>e(t?void 0:w.UNSUPPORTED))}]},mediaCastUnavailable:{get(e,{availability:t=`not-available`}={}){let{media:n}=e;if(!er||!n?.remote?.state)return w.UNSUPPORTED;if(!(t==null||t===`available`))return w.UNAVAILABLE},stateOwnersUpdateHandlers:[(e,t)=>{var n;let{media:r}=t;if(r)return r.disableRemotePlayback||r.hasAttribute(`disableremoteplayback`)||(n=r?.remote)==null||n.watchAvailability(t=>{e({availability:t?`available`:`not-available`})}).catch(t=>{t.name===`NotSupportedError`?e({availability:null}):e({availability:`not-available`})}),()=>{var e;(e=r?.remote)==null||e.cancelWatchAvailability().catch(()=>{})}}]},mediaAirplayUnavailable:{get(e,t){if(!$n)return w.UNSUPPORTED;if(t?.availability===`not-available`)return w.UNAVAILABLE},mediaEvents:[`webkitplaybacktargetavailabilitychanged`],stateOwnersUpdateHandlers:[(e,t)=>{var n;let{media:r}=t;if(r)return r.disableRemotePlayback||r.hasAttribute(`disableremoteplayback`)||(n=r?.remote)==null||n.watchAvailability(t=>{e({availability:t?`available`:`not-available`})}).catch(t=>{t.name===`NotSupportedError`?e({availability:null}):e({availability:`not-available`})}),()=>{var e;(e=r?.remote)==null||e.cancelWatchAvailability().catch(()=>{})}}]},mediaRenditionUnavailable:{get(e){let{media:t}=e;if(!t?.videoRenditions)return w.UNSUPPORTED;if(!t.videoRenditions?.length)return w.UNAVAILABLE},mediaEvents:[`emptied`,`loadstart`],videoRenditionsEvents:[`addrendition`,`removerendition`]},mediaAudioTrackUnavailable:{get(e){let{media:t}=e;if(!t?.audioTracks)return w.UNSUPPORTED;if((t.audioTracks?.length??0)<=1)return w.UNAVAILABLE},mediaEvents:[`emptied`,`loadstart`],audioTracksEvents:[`addtrack`,`removetrack`]},mediaLang:{get(e){let{options:{mediaLang:t}={}}=e;return t??`en`}}},pr={[y.MEDIA_PREVIEW_REQUEST](e,t,{detail:n}){let{media:r}=t,i=n??void 0,a,o;if(r&&i!=null){let[e]=Nn(r,{kind:C.METADATA,label:`thumbnails`}),t=Array.prototype.find.call(e?.cues??[],(e,t,n)=>t===0?e.endTime>i:t===n.length-1?e.startTime<=i:e.startTime<=i&&e.endTime>i);if(t){let e=/'^(?:[a-z]+:)?\/\//i.test(t.text)?void 0:r?.querySelector(`track[label="thumbnails"]`)?.src,n=new URL(t.text,e);o=new URLSearchParams(n.hash).get(`#xywh`).split(`,`).map(e=>+e),a=n.href}}let s=e.mediaDuration.get(t),c=e.mediaChaptersCues.get(t).find((e,t,n)=>t===n.length-1&&s===e.endTime?e.startTime<=i&&e.endTime>=i:e.startTime<=i&&e.endTime>i)?.text;return n!=null&&c==null&&(c=``),{mediaPreviewTime:i,mediaPreviewImage:a,mediaPreviewCoords:o,mediaPreviewChapter:c}},[y.MEDIA_PAUSE_REQUEST](e,t){e.mediaPaused.set(!0,t)},[y.MEDIA_PLAY_REQUEST](e,t){let n=e.mediaStreamType.get(t)===T.LIVE,r=!t.options?.noAutoSeekToLive,i=e.mediaTargetLiveWindow.get(t)>0;if(n&&r&&!i){let n=e.mediaSeekable.get(t)?.[1];if(n){let r=n-(t.options?.seekToLiveOffset??0);e.mediaCurrentTime.set(r,t)}}e.mediaPaused.set(!1,t)},[y.MEDIA_PLAYBACK_RATE_REQUEST](e,t,{detail:n}){let r=n;e.mediaPlaybackRate.set(r,t)},[y.MEDIA_MUTE_REQUEST](e,t){e.mediaMuted.set(!0,t)},[y.MEDIA_UNMUTE_REQUEST](e,t){e.mediaVolume.get(t)||e.mediaVolume.set(.25,t),e.mediaMuted.set(!1,t)},[y.MEDIA_LOOP_REQUEST](e,t,{detail:n}){let r=!!n;return e.mediaLoop.set(r,t),{mediaLoop:r}},[y.MEDIA_VOLUME_REQUEST](e,t,{detail:n}){let r=n;r&&e.mediaMuted.get(t)&&e.mediaMuted.set(!1,t),e.mediaVolume.set(r,t)},[y.MEDIA_SEEK_REQUEST](e,t,{detail:n}){let r=n;e.mediaCurrentTime.set(r,t)},[y.MEDIA_SEEK_TO_LIVE_REQUEST](e,t){let n=e.mediaSeekable.get(t)?.[1];if(Number.isNaN(Number(n)))return;let r=n-(t.options?.seekToLiveOffset??0);e.mediaCurrentTime.set(r,t)},[y.MEDIA_SHOW_SUBTITLES_REQUEST](e,t,{detail:n}){let{options:r}=t,i=tr(t),a=Dn(n),o=a[0]?.language;o&&!r.noSubtitlesLangPref&&O.localStorage.setItem(`media-chrome-pref-subtitles-lang`,o),Mn(Le.SHOWING,i,a)},[y.MEDIA_DISABLE_SUBTITLES_REQUEST](e,t,{detail:n}){let r=tr(t),i=n??[];Mn(Le.DISABLED,r,i)},[y.MEDIA_TOGGLE_SUBTITLES_REQUEST](e,t,{detail:n}){rr(t,n)},[y.MEDIA_RENDITION_REQUEST](e,t,{detail:n}){let r=n;e.mediaRenditionSelected.set(r,t)},[y.MEDIA_AUDIO_TRACK_REQUEST](e,t,{detail:n}){let r=n;e.mediaAudioTrackEnabled.set(r,t)},[y.MEDIA_ENTER_PIP_REQUEST](e,t){e.mediaIsFullscreen.get(t)&&e.mediaIsFullscreen.set(!1,t),e.mediaIsPip.set(!0,t)},[y.MEDIA_EXIT_PIP_REQUEST](e,t){e.mediaIsPip.set(!1,t)},[y.MEDIA_ENTER_FULLSCREEN_REQUEST](e,t,n){e.mediaIsPip.get(t)&&e.mediaIsPip.set(!1,t),e.mediaIsFullscreen.set(!0,t,n)},[y.MEDIA_EXIT_FULLSCREEN_REQUEST](e,t){e.mediaIsFullscreen.set(!1,t)},[y.MEDIA_ENTER_CAST_REQUEST](e,t){e.mediaIsFullscreen.get(t)&&e.mediaIsFullscreen.set(!1,t),e.mediaIsCasting.set(!0,t)},[y.MEDIA_EXIT_CAST_REQUEST](e,t){e.mediaIsCasting.set(!1,t)},[y.MEDIA_AIRPLAY_REQUEST](e,t){e.mediaIsAirplaying.set(!0,t)}},mr=({media:e,fullscreenElement:t,documentElement:n,stateMediator:r=fr,requestMap:i=pr,options:a={},monitorStateOwnersOnlyWithSubscriptions:o=!0})=>{let s=[],c={options:{...a}},l=Object.freeze({mediaPreviewTime:void 0,mediaPreviewImage:void 0,mediaPreviewCoords:void 0,mediaPreviewChapter:void 0}),u=e=>{e!=null&&(ir(e,l)||(l=Object.freeze({...l,...e}),s.forEach(e=>e(l))))},d=()=>{let e=Object.entries(r).reduce((e,[t,{get:n}])=>(e[t]=n(c),e),{});u(e)},f={},p,m=async(e,t)=>{let n=!!p;if(p={...c,...p??{},...e},n)return;await lr(...Object.values(e));let i=s.length>0&&t===0&&o,a=c.media!==p.media,l=c.media?.textTracks!==p.media?.textTracks,m=c.media?.videoRenditions!==p.media?.videoRenditions,h=c.media?.audioTracks!==p.media?.audioTracks,g=c.media?.remote!==p.media?.remote,ee=c.documentElement!==p.documentElement,te=!!c.media&&(a||i),ne=!!c.media?.textTracks&&(l||i),re=!!c.media?.videoRenditions&&(m||i),ie=!!c.media?.audioTracks&&(h||i),ae=!!c.media?.remote&&(g||i),oe=!!c.documentElement&&(ee||i),se=te||ne||re||ie||ae||oe,ce=s.length===0&&t===1&&o,le=!!p.media&&(a||ce),ue=!!p.media?.textTracks&&(l||ce),de=!!p.media?.videoRenditions&&(m||ce),fe=!!p.media?.audioTracks&&(h||ce),pe=!!p.media?.remote&&(g||ce),me=!!p.documentElement&&(ee||ce),he=le||ue||de||fe||pe||me;if(!(se||he)){Object.entries(p).forEach(([e,t])=>{c[e]=t}),d(),p=void 0;return}Object.entries(r).forEach(([e,{get:t,mediaEvents:n=[],textTracksEvents:r=[],videoRenditionsEvents:i=[],audioTracksEvents:a=[],remoteEvents:o=[],rootEvents:s=[],stateOwnersUpdateHandlers:l=[]}])=>{f[e]||(f[e]={});let d=n=>{let r=t(c,n);u({[e]:r})},m;m=f[e].mediaEvents,n.forEach(t=>{m&&te&&(c.media.removeEventListener(t,m),f[e].mediaEvents=void 0),le&&(p.media.addEventListener(t,d),f[e].mediaEvents=d)}),m=f[e].textTracksEvents,r.forEach(t=>{var n,r;m&&ne&&((n=c.media.textTracks)==null||n.removeEventListener(t,m),f[e].textTracksEvents=void 0),ue&&((r=p.media.textTracks)==null||r.addEventListener(t,d),f[e].textTracksEvents=d)}),m=f[e].videoRenditionsEvents,i.forEach(t=>{var n,r;m&&re&&((n=c.media.videoRenditions)==null||n.removeEventListener(t,m),f[e].videoRenditionsEvents=void 0),de&&((r=p.media.videoRenditions)==null||r.addEventListener(t,d),f[e].videoRenditionsEvents=d)}),m=f[e].audioTracksEvents,a.forEach(t=>{var n,r;m&&ie&&((n=c.media.audioTracks)==null||n.removeEventListener(t,m),f[e].audioTracksEvents=void 0),fe&&((r=p.media.audioTracks)==null||r.addEventListener(t,d),f[e].audioTracksEvents=d)}),m=f[e].remoteEvents,o.forEach(t=>{var n,r;m&&ae&&((n=c.media.remote)==null||n.removeEventListener(t,m),f[e].remoteEvents=void 0),pe&&((r=p.media.remote)==null||r.addEventListener(t,d),f[e].remoteEvents=d)}),m=f[e].rootEvents,s.forEach(t=>{m&&oe&&(c.documentElement.removeEventListener(t,m),f[e].rootEvents=void 0),me&&(p.documentElement.addEventListener(t,d),f[e].rootEvents=d)});let h=f[e].stateOwnersUpdateHandlers;if(h&&se&&(Array.isArray(h)?h:[h]).forEach(e=>{typeof e==`function`&&e()}),he){let t=l.map(e=>e(d,p)).filter(e=>typeof e==`function`);f[e].stateOwnersUpdateHandlers=t.length===1?t[0]:t}else se&&(f[e].stateOwnersUpdateHandlers=void 0)}),Object.entries(p).forEach(([e,t])=>{c[e]=t}),d(),p=void 0};return m({media:e,fullscreenElement:t,documentElement:n,options:a}),{dispatch(e){let{type:t,detail:n}=e;if(i[t]&&l.mediaErrorCode==null){u(i[t](r,c,e));return}t===`mediaelementchangerequest`?m({media:n}):t===`fullscreenelementchangerequest`?m({fullscreenElement:n}):t===`documentelementchangerequest`?m({documentElement:n}):t===`optionschangerequest`&&(Object.entries(n??{}).forEach(([e,t])=>{c.options[e]=t}),d())},getState(){return l},subscribe(e){return m({},s.length+1),s.push(e),e(l),()=>{let t=s.indexOf(e);t>=0&&(m({},s.length-1),s.splice(t,1))}}}},hr=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},H=(e,t,n)=>(hr(e,t,`read from private field`),n?n.call(e):t.get(e)),gr=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},_r=(e,t,n,r)=>(hr(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),vr=(e,t,n)=>(hr(e,t,`access private method`),n),yr,br,U,xr,Sr,Cr,wr,Tr,Er,Dr,Or,kr,Ar,jr,Mr,Nr=[`ArrowLeft`,`ArrowRight`,`ArrowUp`,`ArrowDown`,`Enter`,` `,`f`,`m`,`k`,`c`,`l`,`j`,`>`,`<`,`p`],Pr=10,Fr=.025,Ir=.25,Lr=.25,Rr=2,W={DEFAULT_SUBTITLES:`defaultsubtitles`,DEFAULT_STREAM_TYPE:`defaultstreamtype`,DEFAULT_DURATION:`defaultduration`,FULLSCREEN_ELEMENT:`fullscreenelement`,HOTKEYS:`hotkeys`,KEYBOARD_BACKWARD_SEEK_OFFSET:`keyboardbackwardseekoffset`,KEYBOARD_FORWARD_SEEK_OFFSET:`keyboardforwardseekoffset`,KEYBOARD_DOWN_VOLUME_STEP:`keyboarddownvolumestep`,KEYBOARD_UP_VOLUME_STEP:`keyboardupvolumestep`,KEYS_USED:`keysused`,LANG:`lang`,LOOP:`loop`,LIVE_EDGE_OFFSET:`liveedgeoffset`,NO_AUTO_SEEK_TO_LIVE:`noautoseektolive`,NO_DEFAULT_STORE:`nodefaultstore`,NO_HOTKEYS:`nohotkeys`,NO_MUTED_PREF:`nomutedpref`,NO_SUBTITLES_LANG_PREF:`nosubtitleslangpref`,NO_VOLUME_PREF:`novolumepref`,SEEK_TO_LIVE_OFFSET:`seektoliveoffset`},zr=class extends fn{constructor(){super(),gr(this,Er),gr(this,kr),gr(this,jr),this.mediaStateReceivers=[],this.associatedElementSubscriptions=new Map,gr(this,yr,new Cn(this,W.HOTKEYS)),gr(this,br,void 0),gr(this,U,void 0),gr(this,xr,null),gr(this,Sr,void 0),gr(this,Cr,void 0),gr(this,wr,e=>{var t;(t=H(this,U))==null||t.dispatch(e)}),gr(this,Tr,void 0),gr(this,Or,e=>{let{key:t,shiftKey:n}=e;if(!(n&&(t===`/`||t===`?`)||Nr.includes(t))){this.removeEventListener(`keyup`,H(this,Or));return}this.keyboardShortcutHandler(e)}),this.associateElement(this);let e={};_r(this,Sr,t=>{Object.entries(t).forEach(([t,n])=>{if(t in e&&e[t]===n)return;this.propagateMediaState(t,n);let r=t.toLowerCase(),i=new O.CustomEvent(S[r],{composed:!0,detail:n});this.dispatchEvent(i)}),e=t})}static get observedAttributes(){return super.observedAttributes.concat(W.NO_HOTKEYS,W.HOTKEYS,W.DEFAULT_STREAM_TYPE,W.DEFAULT_SUBTITLES,W.DEFAULT_DURATION,W.NO_MUTED_PREF,W.NO_VOLUME_PREF,W.LANG,W.LOOP,W.LIVE_EDGE_OFFSET,W.SEEK_TO_LIVE_OFFSET,W.NO_AUTO_SEEK_TO_LIVE)}get mediaStore(){return H(this,U)}set mediaStore(e){var t;if(H(this,U)&&((t=H(this,Cr))==null||t.call(this),_r(this,Cr,void 0)),_r(this,U,e),!H(this,U)&&!this.hasAttribute(W.NO_DEFAULT_STORE)){vr(this,Er,Dr).call(this);return}_r(this,Cr,H(this,U)?.subscribe(H(this,Sr)))}get fullscreenElement(){return H(this,br)??this}set fullscreenElement(e){var t;this.hasAttribute(W.FULLSCREEN_ELEMENT)&&this.removeAttribute(W.FULLSCREEN_ELEMENT),_r(this,br,e),(t=H(this,U))==null||t.dispatch({type:`fullscreenelementchangerequest`,detail:this.fullscreenElement})}get defaultSubtitles(){return N(this,W.DEFAULT_SUBTITLES)}set defaultSubtitles(e){P(this,W.DEFAULT_SUBTITLES,e)}get defaultStreamType(){return F(this,W.DEFAULT_STREAM_TYPE)}set defaultStreamType(e){I(this,W.DEFAULT_STREAM_TYPE,e)}get defaultDuration(){return j(this,W.DEFAULT_DURATION)}set defaultDuration(e){M(this,W.DEFAULT_DURATION,e)}get noHotkeys(){return N(this,W.NO_HOTKEYS)}set noHotkeys(e){P(this,W.NO_HOTKEYS,e)}get keysUsed(){return F(this,W.KEYS_USED)}set keysUsed(e){I(this,W.KEYS_USED,e)}get liveEdgeOffset(){return j(this,W.LIVE_EDGE_OFFSET)}set liveEdgeOffset(e){M(this,W.LIVE_EDGE_OFFSET,e)}get noAutoSeekToLive(){return N(this,W.NO_AUTO_SEEK_TO_LIVE)}set noAutoSeekToLive(e){P(this,W.NO_AUTO_SEEK_TO_LIVE,e)}get noVolumePref(){return N(this,W.NO_VOLUME_PREF)}set noVolumePref(e){P(this,W.NO_VOLUME_PREF,e)}get noMutedPref(){return N(this,W.NO_MUTED_PREF)}set noMutedPref(e){P(this,W.NO_MUTED_PREF,e)}get noSubtitlesLangPref(){return N(this,W.NO_SUBTITLES_LANG_PREF)}set noSubtitlesLangPref(e){P(this,W.NO_SUBTITLES_LANG_PREF,e)}get noDefaultStore(){return N(this,W.NO_DEFAULT_STORE)}set noDefaultStore(e){P(this,W.NO_DEFAULT_STORE,e)}get resolvedLang(){return Ye()}attributeChangedCallback(e,t,n){var r,i,a,o,s,c,l,u,d,f;if(super.attributeChangedCallback(e,t,n),e===W.NO_HOTKEYS)n!==t&&n===``?(this.hasAttribute(W.HOTKEYS)&&console.warn("Media Chrome: Both `hotkeys` and `nohotkeys` have been set. All hotkeys will be disabled."),this.disableHotkeys()):n!==t&&n===null&&this.enableHotkeys();else if(e===W.HOTKEYS)H(this,yr).value=n;else if(e===W.DEFAULT_SUBTITLES&&n!==t)(r=H(this,U))==null||r.dispatch({type:`optionschangerequest`,detail:{defaultSubtitles:this.hasAttribute(W.DEFAULT_SUBTITLES)}});else if(e===W.DEFAULT_STREAM_TYPE)(i=H(this,U))==null||i.dispatch({type:`optionschangerequest`,detail:{defaultStreamType:this.getAttribute(W.DEFAULT_STREAM_TYPE)??void 0}});else if(e===W.LIVE_EDGE_OFFSET&&n!==t)(a=H(this,U))==null||a.dispatch({type:`optionschangerequest`,detail:{liveEdgeOffset:this.hasAttribute(W.LIVE_EDGE_OFFSET)?+this.getAttribute(W.LIVE_EDGE_OFFSET):void 0,seekToLiveOffset:this.hasAttribute(W.SEEK_TO_LIVE_OFFSET)?+this.getAttribute(W.SEEK_TO_LIVE_OFFSET):this.hasAttribute(W.LIVE_EDGE_OFFSET)?+this.getAttribute(W.LIVE_EDGE_OFFSET):void 0}});else if(e===W.SEEK_TO_LIVE_OFFSET&&n!==t)(o=H(this,U))==null||o.dispatch({type:`optionschangerequest`,detail:{seekToLiveOffset:this.hasAttribute(W.SEEK_TO_LIVE_OFFSET)?+this.getAttribute(W.SEEK_TO_LIVE_OFFSET):this.hasAttribute(W.LIVE_EDGE_OFFSET)?+this.getAttribute(W.LIVE_EDGE_OFFSET):void 0}});else if(e===W.NO_AUTO_SEEK_TO_LIVE)(s=H(this,U))==null||s.dispatch({type:`optionschangerequest`,detail:{noAutoSeekToLive:this.hasAttribute(W.NO_AUTO_SEEK_TO_LIVE)}});else if(e===W.FULLSCREEN_ELEMENT){let e=n?this.getRootNode()?.getElementById(n):void 0;_r(this,br,e),(c=H(this,U))==null||c.dispatch({type:`fullscreenelementchangerequest`,detail:this.fullscreenElement})}else e===W.LANG&&n!==t?(qe(n),(l=H(this,U))==null||l.dispatch({type:`optionschangerequest`,detail:{mediaLang:n}})):e===W.LOOP&&n!==t?(u=H(this,U))==null||u.dispatch({type:y.MEDIA_LOOP_REQUEST,detail:n!=null}):e===W.NO_VOLUME_PREF&&n!==t?(d=H(this,U))==null||d.dispatch({type:`optionschangerequest`,detail:{noVolumePref:this.hasAttribute(W.NO_VOLUME_PREF)}}):e===W.NO_MUTED_PREF&&n!==t&&((f=H(this,U))==null||f.dispatch({type:`optionschangerequest`,detail:{noMutedPref:this.hasAttribute(W.NO_MUTED_PREF)}}))}connectedCallback(){var e,t;this.associateElement(this),!H(this,U)&&!this.hasAttribute(W.NO_DEFAULT_STORE)&&vr(this,Er,Dr).call(this),(e=H(this,U))==null||e.dispatch({type:`documentelementchangerequest`,detail:k}),(t=H(this,U))==null||t.dispatch({type:`fullscreenelementchangerequest`,detail:this.fullscreenElement}),super.connectedCallback(),H(this,U)&&!H(this,Cr)&&_r(this,Cr,H(this,U)?.subscribe(H(this,Sr))),H(this,Tr)!==void 0&&H(this,U)&&this.media&&setTimeout(()=>{var e;this.media?.textTracks?.length&&((e=H(this,U))==null||e.dispatch({type:y.MEDIA_TOGGLE_SUBTITLES_REQUEST,detail:H(this,Tr)}))},0),this.hasAttribute(W.NO_HOTKEYS)?this.disableHotkeys():this.enableHotkeys()}disconnectedCallback(){var e,t,n,r,i;if((e=super.disconnectedCallback)==null||e.call(this),this.disableHotkeys(),H(this,U)){let e=H(this,U).getState();_r(this,Tr,!!e.mediaSubtitlesShowing?.length),(t=H(this,U))==null||t.dispatch({type:`fullscreenelementchangerequest`,detail:void 0}),(n=H(this,U))==null||n.dispatch({type:`documentelementchangerequest`,detail:void 0}),(r=H(this,U))==null||r.dispatch({type:y.MEDIA_TOGGLE_SUBTITLES_REQUEST,detail:!1})}H(this,Cr)&&((i=H(this,Cr))==null||i.call(this),_r(this,Cr,void 0)),this.unassociateElement(this),H(this,xr)&&(H(this,xr).remove(),_r(this,xr,null))}mediaSetCallback(e){var t;super.mediaSetCallback(e),(t=H(this,U))==null||t.dispatch({type:`mediaelementchangerequest`,detail:e}),e.hasAttribute(`tabindex`)||(e.tabIndex=-1)}mediaUnsetCallback(e){var t;super.mediaUnsetCallback(e),(t=H(this,U))==null||t.dispatch({type:`mediaelementchangerequest`,detail:void 0})}propagateMediaState(e,t){Xr(this.mediaStateReceivers,e,t)}associateElement(e){if(!e)return;let{associatedElementSubscriptions:t}=this;if(t.has(e))return;let n=Zr(e,this.registerMediaStateReceiver.bind(this),this.unregisterMediaStateReceiver.bind(this));Object.values(y).forEach(t=>{e.addEventListener(t,H(this,wr))}),t.set(e,n)}unassociateElement(e){if(!e)return;let{associatedElementSubscriptions:t}=this;t.has(e)&&(t.get(e)(),t.delete(e),Object.values(y).forEach(t=>{e.removeEventListener(t,H(this,wr))}))}registerMediaStateReceiver(e){if(!e)return;let t=this.mediaStateReceivers;t.indexOf(e)>-1||(t.push(e),H(this,U)&&Object.entries(H(this,U).getState()).forEach(([t,n])=>{Xr([e],t,n)}))}unregisterMediaStateReceiver(e){let t=this.mediaStateReceivers,n=t.indexOf(e);n<0||t.splice(n,1)}enableHotkeys(){this.addEventListener(`keydown`,vr(this,kr,Ar))}disableHotkeys(){this.removeEventListener(`keydown`,vr(this,kr,Ar)),this.removeEventListener(`keyup`,H(this,Or))}get hotkeys(){return H(this,yr)}set hotkeys(e){I(this,W.HOTKEYS,e)}keyboardShortcutHandler(e){let t=e.target;if((t.getAttribute(W.KEYS_USED)?.split(` `)??t?.keysUsed??[]).map(e=>e===`Space`?` `:e).filter(Boolean).includes(e.key))return;let n,r,i;if(!H(this,yr).contains(`no${e.key.toLowerCase()}`)&&!(e.key===` `&&H(this,yr).contains(`nospace`))&&!(e.shiftKey&&(e.key===`/`||e.key===`?`)&&H(this,yr).contains(`noshift+/`)))switch(e.key){case` `:case`k`:n=H(this,U).getState().mediaPaused?y.MEDIA_PLAY_REQUEST:y.MEDIA_PAUSE_REQUEST,this.dispatchEvent(new O.CustomEvent(n,{composed:!0,bubbles:!0}));break;case`m`:n=this.mediaStore.getState().mediaVolumeLevel===`off`?y.MEDIA_UNMUTE_REQUEST:y.MEDIA_MUTE_REQUEST,this.dispatchEvent(new O.CustomEvent(n,{composed:!0,bubbles:!0}));break;case`f`:n=this.mediaStore.getState().mediaIsFullscreen?y.MEDIA_EXIT_FULLSCREEN_REQUEST:y.MEDIA_ENTER_FULLSCREEN_REQUEST,this.dispatchEvent(new O.CustomEvent(n,{composed:!0,bubbles:!0}));break;case`c`:this.dispatchEvent(new O.CustomEvent(y.MEDIA_TOGGLE_SUBTITLES_REQUEST,{composed:!0,bubbles:!0}));break;case`ArrowLeft`:case`j`:{let e=this.hasAttribute(W.KEYBOARD_BACKWARD_SEEK_OFFSET)?+this.getAttribute(W.KEYBOARD_BACKWARD_SEEK_OFFSET):Pr;r=Math.max((this.mediaStore.getState().mediaCurrentTime??0)-e,0),i=new O.CustomEvent(y.MEDIA_SEEK_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`ArrowRight`:case`l`:{let e=this.hasAttribute(W.KEYBOARD_FORWARD_SEEK_OFFSET)?+this.getAttribute(W.KEYBOARD_FORWARD_SEEK_OFFSET):Pr;r=Math.max((this.mediaStore.getState().mediaCurrentTime??0)+e,0),i=new O.CustomEvent(y.MEDIA_SEEK_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`ArrowUp`:{let e=this.hasAttribute(W.KEYBOARD_UP_VOLUME_STEP)?+this.getAttribute(W.KEYBOARD_UP_VOLUME_STEP):Fr;r=Math.min((this.mediaStore.getState().mediaVolume??1)+e,1),i=new O.CustomEvent(y.MEDIA_VOLUME_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`ArrowDown`:{let e=this.hasAttribute(W.KEYBOARD_DOWN_VOLUME_STEP)?+this.getAttribute(W.KEYBOARD_DOWN_VOLUME_STEP):Fr;r=Math.max((this.mediaStore.getState().mediaVolume??1)-e,0),i=new O.CustomEvent(y.MEDIA_VOLUME_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`<`:{let e=this.mediaStore.getState().mediaPlaybackRate??1;r=Math.max(e-Ir,Lr).toFixed(2),i=new O.CustomEvent(y.MEDIA_PLAYBACK_RATE_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`>`:{let e=this.mediaStore.getState().mediaPlaybackRate??1;r=Math.min(e+Ir,Rr).toFixed(2),i=new O.CustomEvent(y.MEDIA_PLAYBACK_RATE_REQUEST,{composed:!0,bubbles:!0,detail:r}),this.dispatchEvent(i);break}case`/`:case`?`:e.shiftKey&&vr(this,jr,Mr).call(this);break;case`p`:n=this.mediaStore.getState().mediaIsPip?y.MEDIA_EXIT_PIP_REQUEST:y.MEDIA_ENTER_PIP_REQUEST,i=new O.CustomEvent(n,{composed:!0,bubbles:!0}),this.dispatchEvent(i);break;default:break}}};yr=new WeakMap,br=new WeakMap,U=new WeakMap,xr=new WeakMap,Sr=new WeakMap,Cr=new WeakMap,wr=new WeakMap,Tr=new WeakMap,Er=new WeakSet,Dr=function(){this.mediaStore=mr({media:this.media,fullscreenElement:this.fullscreenElement,options:{defaultSubtitles:this.hasAttribute(W.DEFAULT_SUBTITLES),defaultDuration:this.hasAttribute(W.DEFAULT_DURATION)?+this.getAttribute(W.DEFAULT_DURATION):void 0,defaultStreamType:this.getAttribute(W.DEFAULT_STREAM_TYPE)??void 0,liveEdgeOffset:this.hasAttribute(W.LIVE_EDGE_OFFSET)?+this.getAttribute(W.LIVE_EDGE_OFFSET):void 0,seekToLiveOffset:this.hasAttribute(W.SEEK_TO_LIVE_OFFSET)?+this.getAttribute(W.SEEK_TO_LIVE_OFFSET):this.hasAttribute(W.LIVE_EDGE_OFFSET)?+this.getAttribute(W.LIVE_EDGE_OFFSET):void 0,noAutoSeekToLive:this.hasAttribute(W.NO_AUTO_SEEK_TO_LIVE),noVolumePref:this.hasAttribute(W.NO_VOLUME_PREF),noMutedPref:this.hasAttribute(W.NO_MUTED_PREF),noSubtitlesLangPref:this.hasAttribute(W.NO_SUBTITLES_LANG_PREF)}})},Or=new WeakMap,kr=new WeakSet,Ar=function(e){let{metaKey:t,altKey:n,key:r,shiftKey:i}=e,a=i&&(r===`/`||r===`?`);if(a&&H(this,xr)?.open){this.removeEventListener(`keyup`,H(this,Or));return}if(t||n||!a&&!Nr.includes(r)){this.removeEventListener(`keyup`,H(this,Or));return}let o=e.target,s=o instanceof HTMLElement&&(o.tagName.toLowerCase()===`media-volume-range`||o.tagName.toLowerCase()===`media-time-range`);[` `,`ArrowLeft`,`ArrowRight`,`ArrowUp`,`ArrowDown`].includes(r)&&!(H(this,yr).contains(`no${r.toLowerCase()}`)||r===` `&&H(this,yr).contains(`nospace`))&&!s&&e.preventDefault(),this.addEventListener(`keyup`,H(this,Or),{once:!0})},jr=new WeakSet,Mr=function(){H(this,xr)||(_r(this,xr,k.createElement(`media-keyboard-shortcuts-dialog`)),this.appendChild(H(this,xr))),H(this,xr).open=!0};var Br=Object.values(x),Vr=Object.values(Pe),Hr=e=>{var t;let{observedAttributes:n}=e.constructor;!n&&e.nodeName?.includes(`-`)&&(O.customElements.upgrade(e),{observedAttributes:n}=e.constructor);let r=((t=(e?.getAttribute)?.call(e,b.MEDIA_CHROME_ATTRIBUTES))?.split)?.call(t,/\s+/);return Array.isArray(n||r)?(n||r).filter(e=>Br.includes(e)):[]},Ur=e=>(e.nodeName?.includes(`-`)&&O.customElements.get(e.nodeName?.toLowerCase())&&!(e instanceof O.customElements.get(e.nodeName.toLowerCase()))&&O.customElements.upgrade(e),Vr.some(t=>t in e)),Wr=e=>Ur(e)||!!Hr(e).length,Gr=e=>(e?.join)?.call(e,`:`),Kr={[x.MEDIA_SUBTITLES_LIST]:kn,[x.MEDIA_SUBTITLES_SHOWING]:kn,[x.MEDIA_SEEKABLE]:Gr,[x.MEDIA_BUFFERED]:e=>e?.map(Gr).join(` `),[x.MEDIA_PREVIEW_COORDS]:e=>e?.join(` `),[x.MEDIA_RENDITION_LIST]:ze,[x.MEDIA_AUDIO_TRACK_LIST]:Ve},qr=async(e,t,n)=>{if(e.isConnected||await We(0),typeof n==`boolean`||n==null)return P(e,t,n);if(typeof n==`number`)return M(e,t,n);if(typeof n==`string`)return I(e,t,n);if(Array.isArray(n)&&!n.length)return e.removeAttribute(t);let r=Kr[t]?.call(Kr,n)??n;return e.setAttribute(t,r)},Jr=e=>!!e.closest?.call(e,`*[slot="media"]`),Yr=(e,t)=>{if(Jr(e))return;let n=(e,t)=>{Wr(e)&&t(e);let{children:n=[]}=e??{},r=e?.shadowRoot?.children??[];[...n,...r].forEach(e=>Yr(e,t))},r=e?.nodeName.toLowerCase();if(r.includes(`-`)&&!Wr(e)){O.customElements.whenDefined(r).then(()=>{n(e,t)});return}n(e,t)},Xr=(e,t,n)=>{e.forEach(e=>{if(t in e){e[t]=n;return}let r=Hr(e),i=t.toLowerCase();r.includes(i)&&qr(e,i,n)})},Zr=(e,t,n)=>{Yr(e,t);let r=e=>{t(e?.composedPath()[0]??e.target)},i=e=>{n(e?.composedPath()[0]??e.target)};e.addEventListener(y.REGISTER_MEDIA_STATE_RECEIVER,r),e.addEventListener(y.UNREGISTER_MEDIA_STATE_RECEIVER,i);let a=e=>{e.forEach(e=>{let{addedNodes:r=[],removedNodes:i=[],type:a,target:o,attributeName:s}=e;a===`childList`?(Array.prototype.forEach.call(r,e=>Yr(e,t)),Array.prototype.forEach.call(i,e=>Yr(e,n))):a===`attributes`&&s===b.MEDIA_CHROME_ATTRIBUTES&&(Wr(o)?t(o):n(o))})},o=[],s=e=>{let r=e.target;r.name!==`media`&&(o.forEach(e=>Yr(e,n)),o=[...r.assignedElements({flatten:!0})],o.forEach(e=>Yr(e,t)))};e.addEventListener(`slotchange`,s);let c=new MutationObserver(a);return c.observe(e,{childList:!0,attributes:!0,subtree:!0}),()=>{Yr(e,n),e.removeEventListener(`slotchange`,s),c.disconnect(),e.removeEventListener(y.REGISTER_MEDIA_STATE_RECEIVER,r),e.removeEventListener(y.UNREGISTER_MEDIA_STATE_RECEIVER,i)}};O.customElements.get(`media-controller`)||O.customElements.define(`media-controller`,zr);var Qr=zr,$r={PLACEMENT:`placement`,BOUNDS:`bounds`};function ei(e){return`
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
  `}var ti=class extends O.HTMLElement{constructor(){if(super(),this.updateXOffset=()=>{if(!Ct(this,{checkOpacity:!1,checkVisibilityCSS:!1}))return;let e=this.placement;if(e===`left`||e===`right`){this.style.removeProperty(`--media-tooltip-offset-x`);return}let t=getComputedStyle(this),n=bt(this,`#`+this.bounds)??mt(this);if(!n)return;let{x:r,width:i}=n.getBoundingClientRect(),{x:a,width:o}=this.getBoundingClientRect(),s=a+o,c=r+i,l=t.getPropertyValue(`--media-tooltip-offset-x`),u=l?parseFloat(l.replace(`px`,``)):0,d=t.getPropertyValue(`--media-tooltip-container-margin`),f=d?parseFloat(d.replace(`px`,``)):0,p=a-r+u-f,m=s-c+u+f;if(p<0){this.style.setProperty(`--media-tooltip-offset-x`,`${p}px`);return}if(m>0){this.style.setProperty(`--media-tooltip-offset-x`,`${m}px`);return}this.style.removeProperty(`--media-tooltip-offset-x`)},!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}if(this.arrowEl=this.shadowRoot.querySelector(`#arrow`),Object.prototype.hasOwnProperty.call(this,`placement`)){let e=this.placement;delete this.placement,this.placement=e}}static get observedAttributes(){return[$r.PLACEMENT,$r.BOUNDS]}get placement(){return F(this,$r.PLACEMENT)}set placement(e){I(this,$r.PLACEMENT,e)}get bounds(){return F(this,$r.BOUNDS)}set bounds(e){I(this,$r.BOUNDS,e)}};ti.shadowRootOptions={mode:`open`},ti.getTemplateHTML=ei,O.customElements.get(`media-tooltip`)||O.customElements.define(`media-tooltip`,ti);var ni=ti,ri=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},G=(e,t,n)=>(ri(e,t,`read from private field`),n?n.call(e):t.get(e)),ii=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},ai=(e,t,n,r)=>(ri(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),oi=(e,t,n)=>(ri(e,t,`access private method`),n),si,ci,li,ui,di,fi,pi,mi={TOOLTIP_PLACEMENT:`tooltipplacement`,DISABLED:`disabled`,NO_TOOLTIP:`notooltip`};function hi(e,t={}){return`
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
        <template shadowrootmode="${ni.shadowRootOptions.mode}">
          ${ni.getTemplateHTML({})}
        </template>
        <slot name="tooltip-content">
          ${this.getTooltipContentHTML(e)}
        </slot>
      </media-tooltip>
    </slot>
  `}function gi(e,t){return`
    <slot></slot>
  `}function _i(){return``}var K=class extends O.HTMLElement{constructor(){if(super(),ii(this,fi),ii(this,si,void 0),this.preventClick=!1,this.tooltipEl=null,ii(this,ci,e=>{this.preventClick||this.handleClick(e),setTimeout(G(this,li),0)}),ii(this,li,()=>{var e,t;(t=(e=this.tooltipEl)?.updateXOffset)==null||t.call(e)}),ii(this,ui,e=>{let{key:t}=e;if(!this.keysUsed.includes(t)){this.removeEventListener(`keyup`,G(this,ui));return}this.preventClick||this.handleClick(e)}),ii(this,di,e=>{let{metaKey:t,altKey:n,key:r}=e;if(t||n||!this.keysUsed.includes(r)){this.removeEventListener(`keyup`,G(this,ui));return}this.addEventListener(`keyup`,G(this,ui),{once:!0})}),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes),t=this.constructor.getTemplateHTML(e);this.shadowRoot.setHTMLUnsafe?this.shadowRoot.setHTMLUnsafe(t):this.shadowRoot.innerHTML=t}this.tooltipEl=this.shadowRoot.querySelector(`media-tooltip`)}static get observedAttributes(){return[`disabled`,mi.TOOLTIP_PLACEMENT,b.MEDIA_CONTROLLER,x.MEDIA_LANG]}enable(){this.addEventListener(`click`,G(this,ci)),this.addEventListener(`keydown`,G(this,di)),this.tabIndex=0}disable(){this.removeEventListener(`click`,G(this,ci)),this.removeEventListener(`keydown`,G(this,di)),this.removeEventListener(`keyup`,G(this,ui)),this.tabIndex=-1}attributeChangedCallback(e,t,n){var r,i,a,o;e===b.MEDIA_CONTROLLER?(t&&((i=(r=G(this,si))?.unassociateElement)==null||i.call(r,this),ai(this,si,null)),n&&this.isConnected&&(ai(this,si,this.getRootNode()?.getElementById(n)),(o=(a=G(this,si))?.associateElement)==null||o.call(a,this))):e===`disabled`&&n!==t?n==null?this.enable():this.disable():e===mi.TOOLTIP_PLACEMENT&&this.tooltipEl&&n!==t?this.tooltipEl.placement=n:e===x.MEDIA_LANG&&(this.shadowRoot.querySelector(`slot[name="tooltip-content"]`).innerHTML=this.constructor.getTooltipContentHTML()),G(this,li).call(this)}connectedCallback(){var e,t;let{style:n}=A(this.shadowRoot,`:host`);n.setProperty(`display`,`var(--media-control-display, var(--${this.localName}-display, inline-flex))`),this.hasAttribute(`disabled`)?this.disable():this.enable(),this.setAttribute(`role`,`button`);let r=this.getAttribute(b.MEDIA_CONTROLLER);r&&(ai(this,si,this.getRootNode()?.getElementById(r)),(t=(e=G(this,si))?.associateElement)==null||t.call(e,this)),O.customElements.whenDefined(`media-tooltip`).then(()=>oi(this,fi,pi).call(this))}disconnectedCallback(){var e,t;this.disable(),(t=(e=G(this,si))?.unassociateElement)==null||t.call(e,this),ai(this,si,null),this.removeEventListener(`mouseenter`,G(this,li)),this.removeEventListener(`focus`,G(this,li)),this.removeEventListener(`click`,G(this,ci))}get keysUsed(){return[`Enter`,` `]}get tooltipPlacement(){return F(this,mi.TOOLTIP_PLACEMENT)}set tooltipPlacement(e){I(this,mi.TOOLTIP_PLACEMENT,e)}get mediaController(){return F(this,b.MEDIA_CONTROLLER)}set mediaController(e){I(this,b.MEDIA_CONTROLLER,e)}get disabled(){return N(this,mi.DISABLED)}set disabled(e){P(this,mi.DISABLED,e)}get noTooltip(){return N(this,mi.NO_TOOLTIP)}set noTooltip(e){P(this,mi.NO_TOOLTIP,e)}handleClick(e){}};si=new WeakMap,ci=new WeakMap,li=new WeakMap,ui=new WeakMap,di=new WeakMap,fi=new WeakSet,pi=function(){this.addEventListener(`mouseenter`,G(this,li)),this.addEventListener(`focus`,G(this,li)),this.addEventListener(`click`,G(this,ci));let e=this.tooltipPlacement;e&&this.tooltipEl&&(this.tooltipEl.placement=e)},K.shadowRootOptions={mode:`open`},K.getTemplateHTML=hi,K.getSlotTemplateHTML=gi,K.getTooltipContentHTML=_i,O.customElements.get(`media-chrome-button`)||O.customElements.define(`media-chrome-button`,K);var vi=K,yi=`<svg aria-hidden="true" viewBox="0 0 26 24">
  <path d="M22.13 3H3.87a.87.87 0 0 0-.87.87v13.26a.87.87 0 0 0 .87.87h3.4L9 16H5V5h16v11h-4l1.72 2h3.4a.87.87 0 0 0 .87-.87V3.87a.87.87 0 0 0-.86-.87Zm-8.75 11.44a.5.5 0 0 0-.76 0l-4.91 5.73a.5.5 0 0 0 .38.83h9.82a.501.501 0 0 0 .38-.83l-4.91-5.73Z"/>
</svg>
`;function bi(e){return`
    <style>
      :host([${x.MEDIA_IS_AIRPLAYING}]) slot[name=icon] slot:not([name=exit]) {
        display: none !important;
      }

      
      :host(:not([${x.MEDIA_IS_AIRPLAYING}])) slot[name=icon] slot:not([name=enter]) {
        display: none !important;
      }

      :host([${x.MEDIA_IS_AIRPLAYING}]) slot[name=tooltip-enter],
      :host(:not([${x.MEDIA_IS_AIRPLAYING}])) slot[name=tooltip-exit] {
        display: none;
      }
    </style>

    <slot name="icon">
      <slot name="enter">${yi}</slot>
      <slot name="exit">${yi}</slot>
    </slot>
  `}function xi(){return`
    <slot name="tooltip-enter">${D(`start airplay`)}</slot>
    <slot name="tooltip-exit">${D(`stop airplay`)}</slot>
  `}var Si=e=>{let t=e.mediaIsAirplaying?D(`stop airplay`):D(`start airplay`);e.setAttribute(`aria-label`,t)},Ci=class extends K{static get observedAttributes(){return[...super.observedAttributes,x.MEDIA_IS_AIRPLAYING,x.MEDIA_AIRPLAY_UNAVAILABLE]}connectedCallback(){super.connectedCallback(),Si(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===x.MEDIA_IS_AIRPLAYING&&Si(this)}get mediaIsAirplaying(){return N(this,x.MEDIA_IS_AIRPLAYING)}set mediaIsAirplaying(e){P(this,x.MEDIA_IS_AIRPLAYING,e)}get mediaAirplayUnavailable(){return F(this,x.MEDIA_AIRPLAY_UNAVAILABLE)}set mediaAirplayUnavailable(e){I(this,x.MEDIA_AIRPLAY_UNAVAILABLE,e)}handleClick(){let e=new O.CustomEvent(y.MEDIA_AIRPLAY_REQUEST,{composed:!0,bubbles:!0});this.dispatchEvent(e)}};Ci.getSlotTemplateHTML=bi,Ci.getTooltipContentHTML=xi,O.customElements.get(`media-airplay-button`)||O.customElements.define(`media-airplay-button`,Ci);var wi=Ci,Ti=`<svg aria-hidden="true" viewBox="0 0 26 24">
  <path d="M22.83 5.68a2.58 2.58 0 0 0-2.3-2.5c-3.62-.24-11.44-.24-15.06 0a2.58 2.58 0 0 0-2.3 2.5c-.23 4.21-.23 8.43 0 12.64a2.58 2.58 0 0 0 2.3 2.5c3.62.24 11.44.24 15.06 0a2.58 2.58 0 0 0 2.3-2.5c.23-4.21.23-8.43 0-12.64Zm-11.39 9.45a3.07 3.07 0 0 1-1.91.57 3.06 3.06 0 0 1-2.34-1 3.75 3.75 0 0 1-.92-2.67 3.92 3.92 0 0 1 .92-2.77 3.18 3.18 0 0 1 2.43-1 2.94 2.94 0 0 1 2.13.78c.364.359.62.813.74 1.31l-1.43.35a1.49 1.49 0 0 0-1.51-1.17 1.61 1.61 0 0 0-1.29.58 2.79 2.79 0 0 0-.5 1.89 3 3 0 0 0 .49 1.93 1.61 1.61 0 0 0 1.27.58 1.48 1.48 0 0 0 1-.37 2.1 2.1 0 0 0 .59-1.14l1.4.44a3.23 3.23 0 0 1-1.07 1.69Zm7.22 0a3.07 3.07 0 0 1-1.91.57 3.06 3.06 0 0 1-2.34-1 3.75 3.75 0 0 1-.92-2.67 3.88 3.88 0 0 1 .93-2.77 3.14 3.14 0 0 1 2.42-1 3 3 0 0 1 2.16.82 2.8 2.8 0 0 1 .73 1.31l-1.43.35a1.49 1.49 0 0 0-1.51-1.21 1.61 1.61 0 0 0-1.29.58A2.79 2.79 0 0 0 15 12a3 3 0 0 0 .49 1.93 1.61 1.61 0 0 0 1.27.58 1.44 1.44 0 0 0 1-.37 2.1 2.1 0 0 0 .6-1.15l1.4.44a3.17 3.17 0 0 1-1.1 1.7Z"/>
</svg>`,Ei=`<svg aria-hidden="true" viewBox="0 0 26 24">
  <path d="M17.73 14.09a1.4 1.4 0 0 1-1 .37 1.579 1.579 0 0 1-1.27-.58A3 3 0 0 1 15 12a2.8 2.8 0 0 1 .5-1.85 1.63 1.63 0 0 1 1.29-.57 1.47 1.47 0 0 1 1.51 1.2l1.43-.34A2.89 2.89 0 0 0 19 9.07a3 3 0 0 0-2.14-.78 3.14 3.14 0 0 0-2.42 1 3.91 3.91 0 0 0-.93 2.78 3.74 3.74 0 0 0 .92 2.66 3.07 3.07 0 0 0 2.34 1 3.07 3.07 0 0 0 1.91-.57 3.17 3.17 0 0 0 1.07-1.74l-1.4-.45c-.083.43-.3.822-.62 1.12Zm-7.22 0a1.43 1.43 0 0 1-1 .37 1.58 1.58 0 0 1-1.27-.58A3 3 0 0 1 7.76 12a2.8 2.8 0 0 1 .5-1.85 1.63 1.63 0 0 1 1.29-.57 1.47 1.47 0 0 1 1.51 1.2l1.43-.34a2.81 2.81 0 0 0-.74-1.32 2.94 2.94 0 0 0-2.13-.78 3.18 3.18 0 0 0-2.43 1 4 4 0 0 0-.92 2.78 3.74 3.74 0 0 0 .92 2.66 3.07 3.07 0 0 0 2.34 1 3.07 3.07 0 0 0 1.91-.57 3.23 3.23 0 0 0 1.07-1.74l-1.4-.45a2.06 2.06 0 0 1-.6 1.07Zm12.32-8.41a2.59 2.59 0 0 0-2.3-2.51C18.72 3.05 15.86 3 13 3c-2.86 0-5.72.05-7.53.17a2.59 2.59 0 0 0-2.3 2.51c-.23 4.207-.23 8.423 0 12.63a2.57 2.57 0 0 0 2.3 2.5c1.81.13 4.67.19 7.53.19 2.86 0 5.72-.06 7.53-.19a2.57 2.57 0 0 0 2.3-2.5c.23-4.207.23-8.423 0-12.63Zm-1.49 12.53a1.11 1.11 0 0 1-.91 1.11c-1.67.11-4.45.18-7.43.18-2.98 0-5.76-.07-7.43-.18a1.11 1.11 0 0 1-.91-1.11c-.21-4.14-.21-8.29 0-12.43a1.11 1.11 0 0 1 .91-1.11C7.24 4.56 10 4.49 13 4.49s5.76.07 7.43.18a1.11 1.11 0 0 1 .91 1.11c.21 4.14.21 8.29 0 12.43Z"/>
</svg>`;function Di(e){return`
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
      <slot name="on">${Ti}</slot>
      <slot name="off">${Ei}</slot>
    </slot>
  `}function Oi(){return`
    <slot name="tooltip-enable">${D(`Enable captions`)}</slot>
    <slot name="tooltip-disable">${D(`Disable captions`)}</slot>
  `}var ki=e=>{e.setAttribute(`aria-checked`,Pn(e).toString())},Ai=class extends K{static get observedAttributes(){return[...super.observedAttributes,x.MEDIA_SUBTITLES_LIST,x.MEDIA_SUBTITLES_SHOWING]}connectedCallback(){super.connectedCallback(),this.setAttribute(`role`,`button`),this.setAttribute(`aria-label`,D(`closed captions`)),ki(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===x.MEDIA_SUBTITLES_SHOWING&&ki(this)}get mediaSubtitlesList(){return ji(this,x.MEDIA_SUBTITLES_LIST)}set mediaSubtitlesList(e){Mi(this,x.MEDIA_SUBTITLES_LIST,e)}get mediaSubtitlesShowing(){return ji(this,x.MEDIA_SUBTITLES_SHOWING)}set mediaSubtitlesShowing(e){Mi(this,x.MEDIA_SUBTITLES_SHOWING,e)}handleClick(){this.dispatchEvent(new O.CustomEvent(y.MEDIA_TOGGLE_SUBTITLES_REQUEST,{composed:!0,bubbles:!0}))}};Ai.getSlotTemplateHTML=Di,Ai.getTooltipContentHTML=Oi;var ji=(e,t)=>{let n=e.getAttribute(t);return n?En(n):[]},Mi=(e,t,n)=>{if(!n?.length){e.removeAttribute(t);return}let r=kn(n);e.getAttribute(t)!==r&&e.setAttribute(t,r)};O.customElements.get(`media-captions-button`)||O.customElements.define(`media-captions-button`,Ai);var Ni=Ai,Pi=`<svg aria-hidden="true" viewBox="0 0 24 24"><g><path class="cast_caf_icon_arch0" d="M1,18 L1,21 L4,21 C4,19.3 2.66,18 1,18 L1,18 Z"/><path class="cast_caf_icon_arch1" d="M1,14 L1,16 C3.76,16 6,18.2 6,21 L8,21 C8,17.13 4.87,14 1,14 L1,14 Z"/><path class="cast_caf_icon_arch2" d="M1,10 L1,12 C5.97,12 10,16.0 10,21 L12,21 C12,14.92 7.07,10 1,10 L1,10 Z"/><path class="cast_caf_icon_box" d="M21,3 L3,3 C1.9,3 1,3.9 1,5 L1,8 L3,8 L3,5 L21,5 L21,19 L14,19 L14,21 L21,21 C22.1,21 23,20.1 23,19 L23,5 C23,3.9 22.1,3 21,3 L21,3 Z"/></g></svg>`,Fi=`<svg aria-hidden="true" viewBox="0 0 24 24"><g><path class="cast_caf_icon_arch0" d="M1,18 L1,21 L4,21 C4,19.3 2.66,18 1,18 L1,18 Z"/><path class="cast_caf_icon_arch1" d="M1,14 L1,16 C3.76,16 6,18.2 6,21 L8,21 C8,17.13 4.87,14 1,14 L1,14 Z"/><path class="cast_caf_icon_arch2" d="M1,10 L1,12 C5.97,12 10,16.0 10,21 L12,21 C12,14.92 7.07,10 1,10 L1,10 Z"/><path class="cast_caf_icon_box" d="M21,3 L3,3 C1.9,3 1,3.9 1,5 L1,8 L3,8 L3,5 L21,5 L21,19 L14,19 L14,21 L21,21 C22.1,21 23,20.1 23,19 L23,5 C23,3.9 22.1,3 21,3 L21,3 Z"/><path class="cast_caf_icon_boxfill" d="M5,7 L5,8.63 C8,8.6 13.37,14 13.37,17 L19,17 L19,7 Z"/></g></svg>`;function Ii(e){return`
    <style>
      :host([${x.MEDIA_IS_CASTING}]) slot[name=icon] slot:not([name=exit]) {
        display: none !important;
      }

      
      :host(:not([${x.MEDIA_IS_CASTING}])) slot[name=icon] slot:not([name=enter]) {
        display: none !important;
      }

      :host([${x.MEDIA_IS_CASTING}]) slot[name=tooltip-enter],
      :host(:not([${x.MEDIA_IS_CASTING}])) slot[name=tooltip-exit] {
        display: none;
      }
    </style>

    <slot name="icon">
      <slot name="enter">${Pi}</slot>
      <slot name="exit">${Fi}</slot>
    </slot>
  `}function Li(){return`
    <slot name="tooltip-enter">${D(`Start casting`)}</slot>
    <slot name="tooltip-exit">${D(`Stop casting`)}</slot>
  `}var Ri=e=>{let t=e.mediaIsCasting?D(`stop casting`):D(`start casting`);e.setAttribute(`aria-label`,t)},zi=class extends K{static get observedAttributes(){return[...super.observedAttributes,x.MEDIA_IS_CASTING,x.MEDIA_CAST_UNAVAILABLE]}connectedCallback(){super.connectedCallback(),Ri(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===x.MEDIA_IS_CASTING&&Ri(this)}get mediaIsCasting(){return N(this,x.MEDIA_IS_CASTING)}set mediaIsCasting(e){P(this,x.MEDIA_IS_CASTING,e)}get mediaCastUnavailable(){return F(this,x.MEDIA_CAST_UNAVAILABLE)}set mediaCastUnavailable(e){I(this,x.MEDIA_CAST_UNAVAILABLE,e)}handleClick(){let e=this.mediaIsCasting?y.MEDIA_EXIT_CAST_REQUEST:y.MEDIA_ENTER_CAST_REQUEST;this.dispatchEvent(new O.CustomEvent(e,{composed:!0,bubbles:!0}))}};zi.getSlotTemplateHTML=Ii,zi.getTooltipContentHTML=Li,O.customElements.get(`media-cast-button`)||O.customElements.define(`media-cast-button`,zi);var Bi=zi,Vi=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Hi=(e,t,n)=>(Vi(e,t,`read from private field`),n?n.call(e):t.get(e)),Ui=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Wi=(e,t,n,r)=>(Vi(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),Gi=(e,t,n)=>(Vi(e,t,`access private method`),n),Ki,qi,Ji,Yi,Xi,Zi,Qi,$i,ea,ta,na,ra,ia,aa,oa;function sa(e){return`
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
  `}function ca(e){return`
    <slot id="content"></slot>
  `}var la={OPEN:`open`,ANCHOR:`anchor`},ua=class extends O.HTMLElement{constructor(){super(),Ui(this,Yi),Ui(this,Zi),Ui(this,$i),Ui(this,ta),Ui(this,ra),Ui(this,aa),Ui(this,Ki,!1),Ui(this,qi,null),Ui(this,Ji,null)}static get observedAttributes(){return[la.OPEN,la.ANCHOR]}get open(){return N(this,la.OPEN)}set open(e){P(this,la.OPEN,e)}handleEvent(e){switch(e.type){case`invoke`:Gi(this,ta,na).call(this,e);break;case`focusout`:Gi(this,ra,ia).call(this,e);break;case`keydown`:Gi(this,aa,oa).call(this,e);break}}connectedCallback(){Gi(this,Yi,Xi).call(this),this.role||=`dialog`,this.addEventListener(`invoke`,this),this.addEventListener(`focusout`,this),this.addEventListener(`keydown`,this)}disconnectedCallback(){this.removeEventListener(`invoke`,this),this.removeEventListener(`focusout`,this),this.removeEventListener(`keydown`,this)}attributeChangedCallback(e,t,n){Gi(this,Yi,Xi).call(this),e===la.OPEN&&n!==t&&(this.open?Gi(this,Zi,Qi).call(this):Gi(this,$i,ea).call(this))}focus(){Wi(this,qi,xt());let e=!this.dispatchEvent(new Event(`focus`,{composed:!0,cancelable:!0})),t=!this.dispatchEvent(new Event(`focusin`,{composed:!0,bubbles:!0,cancelable:!0}));e||t||this.querySelector(`[autofocus], [tabindex]:not([tabindex="-1"]), [role="menu"]`)?.focus()}get keysUsed(){return[`Escape`,`Tab`]}};Ki=new WeakMap,qi=new WeakMap,Ji=new WeakMap,Yi=new WeakSet,Xi=function(){if(!Hi(this,Ki)&&(Wi(this,Ki,!0),!this.shadowRoot)){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e),queueMicrotask(()=>{let{style:e}=A(this.shadowRoot,`:host`);e.setProperty(`transition`,`display .15s, visibility .15s, opacity .15s ease-in, transform .15s ease-in`)})}},Zi=new WeakSet,Qi=function(){var e;(e=Hi(this,Ji))==null||e.setAttribute(`aria-expanded`,`true`),this.dispatchEvent(new Event(`open`,{composed:!0,bubbles:!0})),this.addEventListener(`transitionend`,()=>this.focus(),{once:!0})},$i=new WeakSet,ea=function(){var e;(e=Hi(this,Ji))==null||e.setAttribute(`aria-expanded`,`false`),this.dispatchEvent(new Event(`close`,{composed:!0,bubbles:!0}))},ta=new WeakSet,na=function(e){Wi(this,Ji,e.relatedTarget),yt(this,e.relatedTarget)||(this.open=!this.open)},ra=new WeakSet,ia=function(e){var t;yt(this,e.relatedTarget)||((t=Hi(this,qi))==null||t.focus(),Hi(this,Ji)&&Hi(this,Ji)!==e.relatedTarget&&this.open&&(this.open=!1))},aa=new WeakSet,oa=function(e){var t,n,r,i,a;let{key:o,ctrlKey:s,altKey:c,metaKey:l}=e;s||c||l||this.keysUsed.includes(o)&&(e.preventDefault(),e.stopPropagation(),o===`Tab`?(e.shiftKey?(n=(t=this.previousElementSibling)?.focus)==null||n.call(t):(i=(r=this.nextElementSibling)?.focus)==null||i.call(r),this.blur()):o===`Escape`&&((a=Hi(this,qi))==null||a.focus(),this.open=!1))},ua.shadowRootOptions={mode:`open`},ua.getTemplateHTML=sa,ua.getSlotTemplateHTML=ca,O.customElements.get(`media-chrome-dialog`)||O.customElements.define(`media-chrome-dialog`,ua);var da=ua,fa=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},q=(e,t,n)=>(fa(e,t,`read from private field`),n?n.call(e):t.get(e)),J=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},pa=(e,t,n,r)=>(fa(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),ma=(e,t,n)=>(fa(e,t,`access private method`),n),ha,ga,_a,va,ya,ba,xa,Sa,Ca,wa,Ta,Ea,Da,Oa,ka,Aa,ja,Ma,Na,Pa,Fa,Ia,La,Ra,za;function Ba(e){return`
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
  `}function Va(e){return``}var Ha=class extends O.HTMLElement{constructor(){if(super(),J(this,wa),J(this,Ea),J(this,Oa),J(this,Aa),J(this,Ma),J(this,Pa),J(this,Ia),J(this,Ra),J(this,ha,void 0),J(this,ga,void 0),J(this,_a,void 0),J(this,va,void 0),J(this,ya,{}),J(this,ba,[]),J(this,xa,()=>{if(this.range.matches(`:focus-visible`)){let{style:e}=A(this.shadowRoot,`:host`);e.setProperty(`--_focus-visible-box-shadow`,`var(--_focus-box-shadow)`)}}),J(this,Sa,()=>{let{style:e}=A(this.shadowRoot,`:host`);e.removeProperty(`--_focus-visible-box-shadow`)}),J(this,Ca,()=>{let e=this.shadowRoot.querySelector(`#segments-clipping`);e&&e.parentNode.append(e)}),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes),t=this.constructor.getTemplateHTML(e);this.shadowRoot.setHTMLUnsafe?this.shadowRoot.setHTMLUnsafe(t):this.shadowRoot.innerHTML=t}this.container=this.shadowRoot.querySelector(`#container`),pa(this,_a,this.shadowRoot.querySelector(`#startpoint`)),pa(this,va,this.shadowRoot.querySelector(`#endpoint`)),this.range=this.shadowRoot.querySelector(`#range`),this.appearance=this.shadowRoot.querySelector(`#appearance`)}static get observedAttributes(){return[`disabled`,`aria-disabled`,b.MEDIA_CONTROLLER]}attributeChangedCallback(e,t,n){var r,i,a,o;e===b.MEDIA_CONTROLLER?(t&&((i=(r=q(this,ha))?.unassociateElement)==null||i.call(r,this),pa(this,ha,null)),n&&this.isConnected&&(pa(this,ha,this.getRootNode()?.getElementById(n)),(o=(a=q(this,ha))?.associateElement)==null||o.call(a,this))):(e===`disabled`||e===`aria-disabled`&&t!==n)&&(n==null?(this.range.removeAttribute(e),ma(this,Ea,Da).call(this)):(this.range.setAttribute(e,n),ma(this,Oa,ka).call(this)))}connectedCallback(){var e,t;let{style:n}=A(this.shadowRoot,`:host`);n.setProperty(`display`,`var(--media-control-display, var(--${this.localName}-display, inline-flex))`),q(this,ya).pointer=A(this.shadowRoot,`#pointer`),q(this,ya).progress=A(this.shadowRoot,`#progress`),q(this,ya).thumb=A(this.shadowRoot,`#thumb, ::slotted([slot="thumb"])`),q(this,ya).activeSegment=A(this.shadowRoot,`#segments-clipping rect:nth-child(0)`);let r=this.getAttribute(b.MEDIA_CONTROLLER);r&&(pa(this,ha,this.getRootNode()?.getElementById(r)),(t=(e=q(this,ha))?.associateElement)==null||t.call(e,this)),this.updateBar(),this.shadowRoot.addEventListener(`focusin`,q(this,xa)),this.shadowRoot.addEventListener(`focusout`,q(this,Sa)),ma(this,Ea,Da).call(this),dt(this.container,q(this,Ca))}disconnectedCallback(){var e,t;ma(this,Oa,ka).call(this),(t=(e=q(this,ha))?.unassociateElement)==null||t.call(e,this),pa(this,ha,null),this.shadowRoot.removeEventListener(`focusin`,q(this,xa)),this.shadowRoot.removeEventListener(`focusout`,q(this,Sa)),ft(this.container,q(this,Ca))}updatePointerBar(e){var t;(t=q(this,ya).pointer)==null||t.style.setProperty(`width`,`${this.getPointerRatio(e)*100}%`)}updateBar(){var e,t;let n=this.range.valueAsNumber*100;(e=q(this,ya).progress)==null||e.style.setProperty(`width`,`${n}%`),(t=q(this,ya).thumb)==null||t.style.setProperty(`left`,`${n}%`)}updateSegments(e){let t=this.shadowRoot.querySelector(`#segments-clipping`);if(t.textContent=``,this.container.classList.toggle(`segments`,!!e?.length),!e?.length)return;let n=[...new Set([+this.range.min,...e.flatMap(e=>[e.start,e.end]),+this.range.max])];pa(this,ba,[...n]);let r=n.pop();for(let[e,i]of n.entries()){let[a,o]=[e===0,e===n.length-1],s=a?`calc(var(--segments-gap) / -1)`:`${i*100}%`,c=`calc(${((o?r:n[e+1])-i)*100}%${a||o?``:` - var(--segments-gap)`})`,l=k.createElementNS(`http://www.w3.org/2000/svg`,`rect`),u=Et(this.shadowRoot,`#segments-clipping rect:nth-child(${e+1})`);u.style.setProperty(`x`,s),u.style.setProperty(`width`,c),t.append(l)}}getPointerRatio(e){return wt(e.clientX,e.clientY,q(this,_a).getBoundingClientRect(),q(this,va).getBoundingClientRect())}get dragging(){return this.hasAttribute(`dragging`)}handleEvent(e){switch(e.type){case`pointermove`:ma(this,Ra,za).call(this,e);break;case`input`:this.updateBar();break;case`pointerenter`:ma(this,Ma,Na).call(this,e);break;case`pointerdown`:ma(this,Aa,ja).call(this,e);break;case`pointerup`:ma(this,Pa,Fa).call(this);break;case`pointerleave`:ma(this,Ia,La).call(this);break}}get keysUsed(){return[`ArrowUp`,`ArrowRight`,`ArrowDown`,`ArrowLeft`]}};ha=new WeakMap,ga=new WeakMap,_a=new WeakMap,va=new WeakMap,ya=new WeakMap,ba=new WeakMap,xa=new WeakMap,Sa=new WeakMap,Ca=new WeakMap,wa=new WeakSet,Ta=function(e){let t=q(this,ya).activeSegment;if(!t)return;let n=this.getPointerRatio(e),r=`#segments-clipping rect:nth-child(${q(this,ba).findIndex((e,t,r)=>{let i=r[t+1];return i!=null&&n>=e&&n<=i})+1})`;(t.selectorText!=r||!t.style.transform)&&(t.selectorText=r,t.style.setProperty(`transform`,`var(--media-range-segment-hover-transform, scaleY(2))`))},Ea=new WeakSet,Da=function(){this.hasAttribute(`disabled`)||!this.isConnected||(this.addEventListener(`input`,this),this.addEventListener(`pointerdown`,this),this.addEventListener(`pointerenter`,this))},Oa=new WeakSet,ka=function(){var e,t;this.removeEventListener(`input`,this),this.removeEventListener(`pointerdown`,this),this.removeEventListener(`pointerenter`,this),this.removeEventListener(`pointerleave`,this),(e=O.window)==null||e.removeEventListener(`pointerup`,this),(t=O.window)==null||t.removeEventListener(`pointermove`,this)},Aa=new WeakSet,ja=function(e){var t;pa(this,ga,e.composedPath().includes(this.range)),(t=O.window)==null||t.addEventListener(`pointerup`,this,{once:!0})},Ma=new WeakSet,Na=function(e){var t;e.pointerType!==`mouse`&&ma(this,Aa,ja).call(this,e),this.addEventListener(`pointerleave`,this,{once:!0}),(t=O.window)==null||t.addEventListener(`pointermove`,this)},Pa=new WeakSet,Fa=function(){var e;(e=O.window)==null||e.removeEventListener(`pointerup`,this),this.toggleAttribute(`dragging`,!1),this.range.disabled=this.hasAttribute(`disabled`)},Ia=new WeakSet,La=function(){var e,t;this.removeEventListener(`pointerleave`,this),(e=O.window)==null||e.removeEventListener(`pointermove`,this),this.toggleAttribute(`dragging`,!1),this.range.disabled=this.hasAttribute(`disabled`),(t=q(this,ya).activeSegment)==null||t.style.removeProperty(`transform`)},Ra=new WeakSet,za=function(e){e.pointerType===`pen`&&e.buttons===0||(this.toggleAttribute(`dragging`,e.buttons===1||e.pointerType!==`mouse`),this.updatePointerBar(e),ma(this,wa,Ta).call(this,e),this.dragging&&(e.pointerType!==`mouse`||!q(this,ga))&&(this.range.disabled=!0,this.range.valueAsNumber=this.getPointerRatio(e),this.range.dispatchEvent(new Event(`input`,{bubbles:!0,composed:!0}))))},Ha.shadowRootOptions={mode:`open`},Ha.getTemplateHTML=Ba,Ha.getContainerTemplateHTML=Va,O.customElements.get(`media-chrome-range`)||O.customElements.define(`media-chrome-range`,Ha);var Ua=Ha,Wa=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Ga=(e,t,n)=>(Wa(e,t,`read from private field`),n?n.call(e):t.get(e)),Ka=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},qa=(e,t,n,r)=>(Wa(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),Ja;function Ya(e){return`
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
  `}var Xa=class extends O.HTMLElement{constructor(){if(super(),Ka(this,Ja,void 0),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}}static get observedAttributes(){return[b.MEDIA_CONTROLLER]}attributeChangedCallback(e,t,n){var r,i,a,o;e===b.MEDIA_CONTROLLER&&(t&&((i=(r=Ga(this,Ja))?.unassociateElement)==null||i.call(r,this),qa(this,Ja,null)),n&&this.isConnected&&(qa(this,Ja,this.getRootNode()?.getElementById(n)),(o=(a=Ga(this,Ja))?.associateElement)==null||o.call(a,this)))}connectedCallback(){var e,t;let n=this.getAttribute(b.MEDIA_CONTROLLER);n&&(qa(this,Ja,this.getRootNode()?.getElementById(n)),(t=(e=Ga(this,Ja))?.associateElement)==null||t.call(e,this))}disconnectedCallback(){var e,t;(t=(e=Ga(this,Ja))?.unassociateElement)==null||t.call(e,this),qa(this,Ja,null)}};Ja=new WeakMap,Xa.shadowRootOptions={mode:`open`},Xa.getTemplateHTML=Ya,O.customElements.get(`media-control-bar`)||O.customElements.define(`media-control-bar`,Xa);var Za=Xa,Qa=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},$a=(e,t,n)=>(Qa(e,t,`read from private field`),n?n.call(e):t.get(e)),eo=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},to=(e,t,n,r)=>(Qa(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),no;function ro(e,t={}){return`
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
  `}function io(e,t){return`
    <slot></slot>
  `}var ao=class extends O.HTMLElement{constructor(){if(super(),eo(this,no,void 0),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}}static get observedAttributes(){return[b.MEDIA_CONTROLLER]}attributeChangedCallback(e,t,n){var r,i,a,o;e===b.MEDIA_CONTROLLER&&(t&&((i=(r=$a(this,no))?.unassociateElement)==null||i.call(r,this),to(this,no,null)),n&&this.isConnected&&(to(this,no,this.getRootNode()?.getElementById(n)),(o=(a=$a(this,no))?.associateElement)==null||o.call(a,this)))}connectedCallback(){var e,t;let{style:n}=A(this.shadowRoot,`:host`);n.setProperty(`display`,`var(--media-control-display, var(--${this.localName}-display, inline-flex))`);let r=this.getAttribute(b.MEDIA_CONTROLLER);r&&(to(this,no,this.getRootNode()?.getElementById(r)),(t=(e=$a(this,no))?.associateElement)==null||t.call(e,this))}disconnectedCallback(){var e,t;(t=(e=$a(this,no))?.unassociateElement)==null||t.call(e,this),to(this,no,null)}};no=new WeakMap,ao.shadowRootOptions={mode:`open`},ao.getTemplateHTML=ro,ao.getSlotTemplateHTML=io,O.customElements.get(`media-text-display`)||O.customElements.define(`media-text-display`,ao);var oo=ao,so=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},co=(e,t,n)=>(so(e,t,`read from private field`),n?n.call(e):t.get(e)),lo=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},uo=(e,t,n,r)=>(so(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),fo;function po(e,t){return`
    <slot>${$e(t.mediaDuration)}</slot>
  `}var mo=class extends ao{constructor(){super(),lo(this,fo,void 0),uo(this,fo,this.shadowRoot.querySelector(`slot`)),co(this,fo).textContent=$e(this.mediaDuration??0)}static get observedAttributes(){return[...super.observedAttributes,x.MEDIA_DURATION]}attributeChangedCallback(e,t,n){e===x.MEDIA_DURATION&&(co(this,fo).textContent=$e(+n)),super.attributeChangedCallback(e,t,n)}get mediaDuration(){return j(this,x.MEDIA_DURATION)}set mediaDuration(e){M(this,x.MEDIA_DURATION,e)}};fo=new WeakMap,mo.getSlotTemplateHTML=po,O.customElements.get(`media-duration-display`)||O.customElements.define(`media-duration-display`,mo);var ho=mo,go={2:D(`Network Error`),3:D(`Decode Error`),4:D(`Source Not Supported`),5:D(`Encryption Error`)},_o={2:D(`A network error caused the media download to fail.`),3:D(`A media error caused playback to be aborted. The media could be corrupt or your browser does not support this format.`),4:D(`An unsupported error occurred. The server or network failed, or your browser does not support this format.`),5:D(`The media is encrypted and there are no keys to decrypt it.`)},vo=e=>e.code===1?null:{title:go[e.code]??`Error ${e.code}`,message:_o[e.code]??e.message},yo=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},bo=(e,t,n)=>(yo(e,t,`read from private field`),n?n.call(e):t.get(e)),xo=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},So=(e,t,n,r)=>(yo(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),Co;function wo(e){return`
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
      ${Eo({code:+e.mediaerrorcode,message:e.mediaerrormessage})}
    </slot>
  `}function To(e){return e.code&&vo(e)!==null}function Eo(e){let{title:t,message:n}=vo(e)??{},r=``;return t&&(r+=`<slot name="error-${e.code}-title"><h3>${t}</h3></slot>`),n&&(r+=`<slot name="error-${e.code}-message"><p>${n}</p></slot>`),r}var Do=[x.MEDIA_ERROR_CODE,x.MEDIA_ERROR_MESSAGE],Oo=class extends ua{constructor(){super(...arguments),xo(this,Co,null)}static get observedAttributes(){return[...super.observedAttributes,...Do]}formatErrorMessage(e){return this.constructor.formatErrorMessage(e)}attributeChangedCallback(e,t,n){if(super.attributeChangedCallback(e,t,n),!Do.includes(e))return;let r=this.mediaError??{code:this.mediaErrorCode,message:this.mediaErrorMessage};if(this.open=To(r),this.open&&(this.shadowRoot.querySelector(`slot`).name=`error-${this.mediaErrorCode}`,this.shadowRoot.querySelector(`#content`).innerHTML=this.formatErrorMessage(r),!this.hasAttribute(`aria-label`))){let{title:e}=vo(r);e&&this.setAttribute(`aria-label`,e)}}get mediaError(){return bo(this,Co)}set mediaError(e){So(this,Co,e)}get mediaErrorCode(){return j(this,`mediaerrorcode`)}set mediaErrorCode(e){M(this,`mediaerrorcode`,e)}get mediaErrorMessage(){return F(this,`mediaerrormessage`)}set mediaErrorMessage(e){I(this,`mediaerrormessage`,e)}};Co=new WeakMap,Oo.getSlotTemplateHTML=wo,Oo.formatErrorMessage=Eo,O.customElements.get(`media-error-dialog`)||O.customElements.define(`media-error-dialog`,Oo);var ko=Oo,Ao=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},jo=(e,t,n)=>(Ao(e,t,`read from private field`),n?n.call(e):t.get(e)),Mo=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},No,Po;function Fo(e){return`
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
      ${Io()}
    </slot>
  `}function Io(){return`
    <h2>Keyboard Shortcuts</h2>
    <table class="shortcuts-table">${[{keys:[`Space`,`k`],description:`Toggle Playback`},{keys:[`m`],description:`Toggle mute`},{keys:[`f`],description:`Toggle fullscreen`},{keys:[`c`],description:`Toggle captions or subtitles, if available`},{keys:[`p`],description:`Toggle Picture in Picture`},{keys:[`←`,`j`],description:`Seek back 10s`},{keys:[`→`,`l`],description:`Seek forward 10s`},{keys:[`↑`],description:`Turn volume up`},{keys:[`↓`],description:`Turn volume down`},{keys:[`< (SHIFT+,)`],description:`Decrease playback rate`},{keys:[`> (SHIFT+.)`],description:`Increase playback rate`}].map(({keys:e,description:t})=>`
      <tr>
        <td>
          <div class="key-combo">${e.map((e,t)=>t>0?`<span class="key-separator">or</span><span class="key">${e}</span>`:`<span class="key">${e}</span>`).join(``)}</div>
        </td>
        <td class="description">${t}</td>
      </tr>
    `).join(``)}</table>
  `}var Lo=class extends ua{constructor(){super(...arguments),Mo(this,No,e=>{if(!this.open)return;let t=this.shadowRoot?.querySelector(`#content`);if(!t)return;let n=e.composedPath(),r=n[0]===this||n.includes(this),i=n.includes(t);r&&!i&&(this.open=!1)}),Mo(this,Po,e=>{if(!this.open)return;let t=e.shiftKey&&(e.key===`/`||e.key===`?`);(e.key===`Escape`||t)&&!e.ctrlKey&&!e.altKey&&!e.metaKey&&(this.open=!1,e.preventDefault(),e.stopPropagation())})}connectedCallback(){super.connectedCallback(),this.open&&(this.addEventListener(`click`,jo(this,No)),document.addEventListener(`keydown`,jo(this,Po)))}disconnectedCallback(){this.removeEventListener(`click`,jo(this,No)),document.removeEventListener(`keydown`,jo(this,Po))}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===`open`&&(this.open?(this.addEventListener(`click`,jo(this,No)),document.addEventListener(`keydown`,jo(this,Po))):(this.removeEventListener(`click`,jo(this,No)),document.removeEventListener(`keydown`,jo(this,Po))))}};No=new WeakMap,Po=new WeakMap,Lo.getSlotTemplateHTML=Fo,O.customElements.get(`media-keyboard-shortcuts-dialog`)||O.customElements.define(`media-keyboard-shortcuts-dialog`,Lo);var Ro=Lo,zo=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Bo=(e,t,n)=>(zo(e,t,`read from private field`),n?n.call(e):t.get(e)),Vo=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Ho=(e,t,n,r)=>(zo(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),Uo,Wo=`<svg aria-hidden="true" viewBox="0 0 26 24">
  <path d="M16 3v2.5h3.5V9H22V3h-6ZM4 9h2.5V5.5H10V3H4v6Zm15.5 9.5H16V21h6v-6h-2.5v3.5ZM6.5 15H4v6h6v-2.5H6.5V15Z"/>
</svg>`,Go=`<svg aria-hidden="true" viewBox="0 0 26 24">
  <path d="M18.5 6.5V3H16v6h6V6.5h-3.5ZM16 21h2.5v-3.5H22V15h-6v6ZM4 17.5h3.5V21H10v-6H4v2.5Zm3.5-11H4V9h6V3H7.5v3.5Z"/>
</svg>`;function Ko(e){return`
    <style>
      :host([${x.MEDIA_IS_FULLSCREEN}]) slot[name=icon] slot:not([name=exit]) {
        display: none !important;
      }

      
      :host(:not([${x.MEDIA_IS_FULLSCREEN}])) slot[name=icon] slot:not([name=enter]) {
        display: none !important;
      }

      :host([${x.MEDIA_IS_FULLSCREEN}]) slot[name=tooltip-enter],
      :host(:not([${x.MEDIA_IS_FULLSCREEN}])) slot[name=tooltip-exit] {
        display: none;
      }
    </style>

    <slot name="icon">
      <slot name="enter">${Wo}</slot>
      <slot name="exit">${Go}</slot>
    </slot>
  `}function qo(){return`
    <slot name="tooltip-enter">${D(`Enter fullscreen mode`)}</slot>
    <slot name="tooltip-exit">${D(`Exit fullscreen mode`)}</slot>
  `}var Jo=e=>{let t=e.mediaIsFullscreen?D(`exit fullscreen mode`):D(`enter fullscreen mode`);e.setAttribute(`aria-label`,t)},Yo=class extends K{constructor(){super(...arguments),Vo(this,Uo,null)}static get observedAttributes(){return[...super.observedAttributes,x.MEDIA_IS_FULLSCREEN,x.MEDIA_FULLSCREEN_UNAVAILABLE]}connectedCallback(){super.connectedCallback(),Jo(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===x.MEDIA_IS_FULLSCREEN&&Jo(this)}get mediaFullscreenUnavailable(){return F(this,x.MEDIA_FULLSCREEN_UNAVAILABLE)}set mediaFullscreenUnavailable(e){I(this,x.MEDIA_FULLSCREEN_UNAVAILABLE,e)}get mediaIsFullscreen(){return N(this,x.MEDIA_IS_FULLSCREEN)}set mediaIsFullscreen(e){P(this,x.MEDIA_IS_FULLSCREEN,e)}handleClick(e){Ho(this,Uo,e);let t=Bo(this,Uo)instanceof PointerEvent,n=this.mediaIsFullscreen?new O.CustomEvent(y.MEDIA_EXIT_FULLSCREEN_REQUEST,{composed:!0,bubbles:!0}):new O.CustomEvent(y.MEDIA_ENTER_FULLSCREEN_REQUEST,{composed:!0,bubbles:!0,detail:t});this.dispatchEvent(n)}};Uo=new WeakMap,Yo.getSlotTemplateHTML=Ko,Yo.getTooltipContentHTML=qo,O.customElements.get(`media-fullscreen-button`)||O.customElements.define(`media-fullscreen-button`,Yo);var Xo=Yo,{MEDIA_TIME_IS_LIVE:Zo,MEDIA_PAUSED:Qo}=x,{MEDIA_SEEK_TO_LIVE_REQUEST:$o,MEDIA_PLAY_REQUEST:es}=y,ts=`<svg viewBox="0 0 6 12" aria-hidden="true"><circle cx="3" cy="6" r="2"></circle></svg>`;function ns(e){return`
    <style>
      :host { --media-tooltip-display: none; }
      
      slot[name=indicator] > *,
      :host ::slotted([slot=indicator]) {
        
        min-width: auto;
        fill: var(--media-live-button-icon-color, rgb(140, 140, 140));
        color: var(--media-live-button-icon-color, rgb(140, 140, 140));
      }

      :host([${Zo}]:not([${Qo}])) slot[name=indicator] > *,
      :host([${Zo}]:not([${Qo}])) ::slotted([slot=indicator]) {
        fill: var(--media-live-button-indicator-color, rgb(255, 0, 0));
        color: var(--media-live-button-indicator-color, rgb(255, 0, 0));
      }

      :host([${Zo}]:not([${Qo}])) {
        cursor: var(--media-cursor, not-allowed);
      }

      slot[name=text]{
        text-transform: uppercase;
      }

    </style>

    <slot name="indicator">${ts}</slot>
    
    <slot name="spacer">&nbsp;</slot><slot name="text">${D(`live`)}</slot>
  `}var rs=e=>{let t=e.mediaPaused||!e.mediaTimeIsLive,n=D(t?`seek to live`:`playing live`);e.setAttribute(`aria-label`,n);let r=e.shadowRoot?.querySelector(`slot[name="text"]`);r&&(r.textContent=D(`live`)),t?e.removeAttribute(`aria-disabled`):e.setAttribute(`aria-disabled`,`true`)},is=class extends K{static get observedAttributes(){return[...super.observedAttributes,Zo,Qo]}connectedCallback(){super.connectedCallback(),rs(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),rs(this)}get mediaPaused(){return N(this,x.MEDIA_PAUSED)}set mediaPaused(e){P(this,x.MEDIA_PAUSED,e)}get mediaTimeIsLive(){return N(this,x.MEDIA_TIME_IS_LIVE)}set mediaTimeIsLive(e){P(this,x.MEDIA_TIME_IS_LIVE,e)}handleClick(){!this.mediaPaused&&this.mediaTimeIsLive||(this.dispatchEvent(new O.CustomEvent($o,{composed:!0,bubbles:!0})),this.hasAttribute(Qo)&&this.dispatchEvent(new O.CustomEvent(es,{composed:!0,bubbles:!0})))}};is.getSlotTemplateHTML=ns,O.customElements.get(`media-live-button`)||O.customElements.define(`media-live-button`,is);var as=is,os=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},ss=(e,t,n)=>(os(e,t,`read from private field`),n?n.call(e):t.get(e)),cs=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},ls=(e,t,n,r)=>(os(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),us,ds,fs={LOADING_DELAY:`loadingdelay`,NO_AUTOHIDE:`noautohide`},ps=500,ms=`
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
`;function hs(e){return`
    <style>
      :host {
        display: var(--media-control-display, var(--media-loading-indicator-display, inline-block));
        vertical-align: middle;
        box-sizing: border-box;
        --_loading-indicator-delay: var(--media-loading-indicator-transition-delay, ${ps}ms);
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

      :host([${x.MEDIA_LOADING}]:not([${x.MEDIA_PAUSED}])) slot[name=icon] > *,
      :host([${x.MEDIA_LOADING}]:not([${x.MEDIA_PAUSED}])) ::slotted([slot=icon]) {
        opacity: var(--media-loading-indicator-opacity, 1);
        transition: opacity 0.15s var(--_loading-indicator-delay);
      }

      :host #status {
        visibility: var(--media-loading-indicator-opacity, hidden);
        transition: visibility 0.15s;
      }

      :host([${x.MEDIA_LOADING}]:not([${x.MEDIA_PAUSED}])) #status {
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

    <slot name="icon">${ms}</slot>
    <div id="status" role="status" aria-live="polite">${D(`media loading`)}</div>
  `}var gs=class extends O.HTMLElement{constructor(){if(super(),cs(this,us,void 0),cs(this,ds,ps),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}}static get observedAttributes(){return[b.MEDIA_CONTROLLER,x.MEDIA_PAUSED,x.MEDIA_LOADING,fs.LOADING_DELAY]}attributeChangedCallback(e,t,n){var r,i,a,o;e===fs.LOADING_DELAY&&t!==n?this.loadingDelay=Number(n):e===b.MEDIA_CONTROLLER&&(t&&((i=(r=ss(this,us))?.unassociateElement)==null||i.call(r,this),ls(this,us,null)),n&&this.isConnected&&(ls(this,us,this.getRootNode()?.getElementById(n)),(o=(a=ss(this,us))?.associateElement)==null||o.call(a,this)))}connectedCallback(){var e,t;let n=this.getAttribute(b.MEDIA_CONTROLLER);n&&(ls(this,us,this.getRootNode()?.getElementById(n)),(t=(e=ss(this,us))?.associateElement)==null||t.call(e,this))}disconnectedCallback(){var e,t;(t=(e=ss(this,us))?.unassociateElement)==null||t.call(e,this),ls(this,us,null)}get loadingDelay(){return ss(this,ds)}set loadingDelay(e){ls(this,ds,e);let{style:t}=A(this.shadowRoot,`:host`);t.setProperty(`--_loading-indicator-delay`,`var(--media-loading-indicator-transition-delay, ${e}ms)`)}get mediaPaused(){return N(this,x.MEDIA_PAUSED)}set mediaPaused(e){P(this,x.MEDIA_PAUSED,e)}get mediaLoading(){return N(this,x.MEDIA_LOADING)}set mediaLoading(e){P(this,x.MEDIA_LOADING,e)}get mediaController(){return F(this,b.MEDIA_CONTROLLER)}set mediaController(e){I(this,b.MEDIA_CONTROLLER,e)}get noAutohide(){return N(this,fs.NO_AUTOHIDE)}set noAutohide(e){P(this,fs.NO_AUTOHIDE,e)}};us=new WeakMap,ds=new WeakMap,gs.shadowRootOptions={mode:`open`},gs.getTemplateHTML=hs,O.customElements.get(`media-loading-indicator`)||O.customElements.define(`media-loading-indicator`,gs);var _s=gs,vs=`<svg aria-hidden="true" viewBox="0 0 24 24">
  <path d="M16.5 12A4.5 4.5 0 0 0 14 8v2.18l2.45 2.45a4.22 4.22 0 0 0 .05-.63Zm2.5 0a6.84 6.84 0 0 1-.54 2.64L20 16.15A8.8 8.8 0 0 0 21 12a9 9 0 0 0-7-8.77v2.06A7 7 0 0 1 19 12ZM4.27 3 3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25A6.92 6.92 0 0 1 14 18.7v2.06A9 9 0 0 0 17.69 19l2 2.05L21 19.73l-9-9L4.27 3ZM12 4 9.91 6.09 12 8.18V4Z"/>
</svg>`,ys=`<svg aria-hidden="true" viewBox="0 0 24 24">
  <path d="M3 9v6h4l5 5V4L7 9H3Zm13.5 3A4.5 4.5 0 0 0 14 8v8a4.47 4.47 0 0 0 2.5-4Z"/>
</svg>`,bs=`<svg aria-hidden="true" viewBox="0 0 24 24">
  <path d="M3 9v6h4l5 5V4L7 9H3Zm13.5 3A4.5 4.5 0 0 0 14 8v8a4.47 4.47 0 0 0 2.5-4ZM14 3.23v2.06a7 7 0 0 1 0 13.42v2.06a9 9 0 0 0 0-17.54Z"/>
</svg>`;function xs(e){return`
    <style>
      :host(:not([${x.MEDIA_VOLUME_LEVEL}])) slot[name=icon] slot:not([name=high]),
      :host([${x.MEDIA_VOLUME_LEVEL}=high]) slot[name=icon] slot:not([name=high]) {
        display: none !important;
      }

      :host([${x.MEDIA_VOLUME_LEVEL}=off]) slot[name=icon] slot:not([name=off]) {
        display: none !important;
      }

      :host([${x.MEDIA_VOLUME_LEVEL}=low]) slot[name=icon] slot:not([name=low]) {
        display: none !important;
      }

      :host([${x.MEDIA_VOLUME_LEVEL}=medium]) slot[name=icon] slot:not([name=medium]) {
        display: none !important;
      }

      :host(:not([${x.MEDIA_VOLUME_LEVEL}=off])) slot[name=tooltip-unmute],
      :host([${x.MEDIA_VOLUME_LEVEL}=off]) slot[name=tooltip-mute] {
        display: none;
      }
    </style>

    <slot name="icon">
      <slot name="off">${vs}</slot>
      <slot name="low">${ys}</slot>
      <slot name="medium">${ys}</slot>
      <slot name="high">${bs}</slot>
    </slot>
  `}function Ss(){return`
    <slot name="tooltip-mute">${D(`Mute`)}</slot>
    <slot name="tooltip-unmute">${D(`Unmute`)}</slot>
  `}var Cs=e=>{let t=e.mediaVolumeLevel===`off`?D(`unmute`):D(`mute`);e.setAttribute(`aria-label`,t)},ws=class extends K{static get observedAttributes(){return[...super.observedAttributes,x.MEDIA_VOLUME_LEVEL]}connectedCallback(){super.connectedCallback(),Cs(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===x.MEDIA_VOLUME_LEVEL&&Cs(this)}get mediaVolumeLevel(){return F(this,x.MEDIA_VOLUME_LEVEL)}set mediaVolumeLevel(e){I(this,x.MEDIA_VOLUME_LEVEL,e)}handleClick(){let e=this.mediaVolumeLevel===`off`?y.MEDIA_UNMUTE_REQUEST:y.MEDIA_MUTE_REQUEST;this.dispatchEvent(new O.CustomEvent(e,{composed:!0,bubbles:!0}))}};ws.getSlotTemplateHTML=xs,ws.getTooltipContentHTML=Ss,O.customElements.get(`media-mute-button`)||O.customElements.define(`media-mute-button`,ws);var Ts=ws,Es=`<svg aria-hidden="true" viewBox="0 0 28 24">
  <path d="M24 3H4a1 1 0 0 0-1 1v16a1 1 0 0 0 1 1h20a1 1 0 0 0 1-1V4a1 1 0 0 0-1-1Zm-1 16H5V5h18v14Zm-3-8h-7v5h7v-5Z"/>
</svg>`;function Ds(e){return`
    <style>
      :host([${x.MEDIA_IS_PIP}]) slot[name=icon] slot:not([name=exit]) {
        display: none !important;
      }

      :host(:not([${x.MEDIA_IS_PIP}])) slot[name=icon] slot:not([name=enter]) {
        display: none !important;
      }

      :host([${x.MEDIA_IS_PIP}]) slot[name=tooltip-enter],
      :host(:not([${x.MEDIA_IS_PIP}])) slot[name=tooltip-exit] {
        display: none;
      }
    </style>

    <slot name="icon">
      <slot name="enter">${Es}</slot>
      <slot name="exit">${Es}</slot>
    </slot>
  `}function Os(){return`
    <slot name="tooltip-enter">${D(`Enter picture in picture mode`)}</slot>
    <slot name="tooltip-exit">${D(`Exit picture in picture mode`)}</slot>
  `}var ks=e=>{let t=e.mediaIsPip?D(`exit picture in picture mode`):D(`enter picture in picture mode`);e.setAttribute(`aria-label`,t)},As=class extends K{static get observedAttributes(){return[...super.observedAttributes,x.MEDIA_IS_PIP,x.MEDIA_PIP_UNAVAILABLE]}connectedCallback(){super.connectedCallback(),ks(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===x.MEDIA_IS_PIP&&ks(this)}get mediaPipUnavailable(){return F(this,x.MEDIA_PIP_UNAVAILABLE)}set mediaPipUnavailable(e){I(this,x.MEDIA_PIP_UNAVAILABLE,e)}get mediaIsPip(){return N(this,x.MEDIA_IS_PIP)}set mediaIsPip(e){P(this,x.MEDIA_IS_PIP,e)}handleClick(){let e=this.mediaIsPip?y.MEDIA_EXIT_PIP_REQUEST:y.MEDIA_ENTER_PIP_REQUEST;this.dispatchEvent(new O.CustomEvent(e,{composed:!0,bubbles:!0}))}};As.getSlotTemplateHTML=Ds,As.getTooltipContentHTML=Os,O.customElements.get(`media-pip-button`)||O.customElements.define(`media-pip-button`,As);var js=As,Ms=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Ns=(e,t,n)=>(Ms(e,t,`read from private field`),n?n.call(e):t.get(e)),Ps=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Fs,Is={RATES:`rates`},Ls=[1,1.2,1.5,1.7,2];function Rs(e){return Math.round(e*100)/100}function zs(e){return`
    <style>
      :host {
        min-width: 5ch;
        padding: var(--media-button-padding, var(--media-control-padding, 10px 5px));
      }
    </style>
    <slot name="icon">${e.mediaplaybackrate?Rs(+e.mediaplaybackrate):1}x</slot>
  `}function Bs(){return D(`Playback rate`)}var Vs=class extends K{constructor(){super(),Ps(this,Fs,new Cn(this,Is.RATES,{defaultValue:Ls})),this.container=this.shadowRoot.querySelector(`slot[name="icon"]`),this.container.innerHTML=`${Rs(this.mediaPlaybackRate??1)}x`}static get observedAttributes(){return[...super.observedAttributes,x.MEDIA_PLAYBACK_RATE,Is.RATES]}attributeChangedCallback(e,t,n){if(super.attributeChangedCallback(e,t,n),e===Is.RATES&&(Ns(this,Fs).value=n),e===x.MEDIA_PLAYBACK_RATE){let e=n?+n:NaN,t=Rs(Number.isNaN(e)?1:e);this.container.innerHTML=`${t}x`,this.setAttribute(`aria-label`,D(`Playback rate {playbackRate}`,{playbackRate:t}))}}get rates(){return Ns(this,Fs)}set rates(e){e?Array.isArray(e)?Ns(this,Fs).value=e.join(` `):typeof e==`string`&&(Ns(this,Fs).value=e):Ns(this,Fs).value=``}get mediaPlaybackRate(){return j(this,x.MEDIA_PLAYBACK_RATE,1)}set mediaPlaybackRate(e){M(this,x.MEDIA_PLAYBACK_RATE,e)}handleClick(){let e=Array.from(Ns(this,Fs).values(),e=>+e).sort((e,t)=>e-t),t=e.find(e=>e>this.mediaPlaybackRate)??e[0]??1,n=new O.CustomEvent(y.MEDIA_PLAYBACK_RATE_REQUEST,{composed:!0,bubbles:!0,detail:t});this.dispatchEvent(n)}};Fs=new WeakMap,Vs.getSlotTemplateHTML=zs,Vs.getTooltipContentHTML=Bs,O.customElements.get(`media-playback-rate-button`)||O.customElements.define(`media-playback-rate-button`,Vs);var Hs=Vs,Us=`<svg aria-hidden="true" viewBox="0 0 24 24">
  <path d="m6 21 15-9L6 3v18Z"/>
</svg>`,Ws=`<svg aria-hidden="true" viewBox="0 0 24 24">
  <path d="M6 20h4V4H6v16Zm8-16v16h4V4h-4Z"/>
</svg>`;function Gs(e){return`
    <style>
      :host([${x.MEDIA_PAUSED}]) slot[name=pause],
      :host(:not([${x.MEDIA_PAUSED}])) slot[name=play] {
        display: none !important;
      }

      :host([${x.MEDIA_PAUSED}]) slot[name=tooltip-pause],
      :host(:not([${x.MEDIA_PAUSED}])) slot[name=tooltip-play] {
        display: none;
      }
    </style>

    <slot name="icon">
      <slot name="play">${Us}</slot>
      <slot name="pause">${Ws}</slot>
    </slot>
  `}function Ks(){return`
    <slot name="tooltip-play">${D(`Play`)}</slot>
    <slot name="tooltip-pause">${D(`Pause`)}</slot>
  `}var qs=e=>{let t=e.mediaPaused?D(`play`):D(`pause`);e.setAttribute(`aria-label`,t)},Js=class extends K{static get observedAttributes(){return[...super.observedAttributes,x.MEDIA_PAUSED,x.MEDIA_ENDED]}connectedCallback(){super.connectedCallback(),qs(this)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),(e===x.MEDIA_PAUSED||e===x.MEDIA_LANG)&&qs(this)}get mediaPaused(){return N(this,x.MEDIA_PAUSED)}set mediaPaused(e){P(this,x.MEDIA_PAUSED,e)}handleClick(){let e=this.mediaPaused?y.MEDIA_PLAY_REQUEST:y.MEDIA_PAUSE_REQUEST;this.dispatchEvent(new O.CustomEvent(e,{composed:!0,bubbles:!0}))}};Js.getSlotTemplateHTML=Gs,Js.getTooltipContentHTML=Ks,O.customElements.get(`media-play-button`)||O.customElements.define(`media-play-button`,Js);var Ys=Js,Xs={PLACEHOLDER_SRC:`placeholdersrc`,SRC:`src`};function Zs(e){return`
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
  `}var Qs=e=>{e.style.removeProperty(`background-image`)},$s=(e,t)=>{e.style[`background-image`]=`url('${t}')`},ec=class extends O.HTMLElement{static get observedAttributes(){return[Xs.PLACEHOLDER_SRC,Xs.SRC]}constructor(){if(super(),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}this.image=this.shadowRoot.querySelector(`#image`)}attributeChangedCallback(e,t,n){e===Xs.SRC&&(n==null?this.image.removeAttribute(Xs.SRC):this.image.setAttribute(Xs.SRC,n)),e===Xs.PLACEHOLDER_SRC&&(n==null?Qs(this.image):$s(this.image,n))}get placeholderSrc(){return F(this,Xs.PLACEHOLDER_SRC)}set placeholderSrc(e){I(this,Xs.SRC,e)}get src(){return F(this,Xs.SRC)}set src(e){I(this,Xs.SRC,e)}};ec.shadowRootOptions={mode:`open`},ec.getTemplateHTML=Zs,O.customElements.get(`media-poster-image`)||O.customElements.define(`media-poster-image`,ec);var tc=ec,nc=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},rc=(e,t,n)=>(nc(e,t,`read from private field`),n?n.call(e):t.get(e)),ic=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},ac=(e,t,n,r)=>(nc(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),oc,sc=class extends ao{constructor(){super(),ic(this,oc,void 0),ac(this,oc,this.shadowRoot.querySelector(`slot`))}static get observedAttributes(){return[...super.observedAttributes,x.MEDIA_PREVIEW_CHAPTER,x.MEDIA_LANG]}attributeChangedCallback(e,t,n){if(super.attributeChangedCallback(e,t,n),(e===x.MEDIA_PREVIEW_CHAPTER||e===x.MEDIA_LANG)&&n!==t&&n!=null)if(rc(this,oc).textContent=n,n!==``){let e=D(`chapter: {chapterName}`,{chapterName:n});this.setAttribute(`aria-valuetext`,e)}else this.removeAttribute(`aria-valuetext`)}get mediaPreviewChapter(){return F(this,x.MEDIA_PREVIEW_CHAPTER)}set mediaPreviewChapter(e){I(this,x.MEDIA_PREVIEW_CHAPTER,e)}};oc=new WeakMap,O.customElements.get(`media-preview-chapter-display`)||O.customElements.define(`media-preview-chapter-display`,sc);var cc=sc,lc=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},uc=(e,t,n)=>(lc(e,t,`read from private field`),n?n.call(e):t.get(e)),dc=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},fc=(e,t,n,r)=>(lc(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),pc;function mc(e){return`
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
  `}var hc=class extends O.HTMLElement{constructor(){if(super(),dc(this,pc,void 0),!this.shadowRoot){this.attachShadow(this.constructor.shadowRootOptions);let e=pt(this.attributes);this.shadowRoot.innerHTML=this.constructor.getTemplateHTML(e)}}static get observedAttributes(){return[b.MEDIA_CONTROLLER,x.MEDIA_PREVIEW_IMAGE,x.MEDIA_PREVIEW_COORDS]}connectedCallback(){var e,t;let n=this.getAttribute(b.MEDIA_CONTROLLER);n&&(fc(this,pc,this.getRootNode()?.getElementById(n)),(t=(e=uc(this,pc))?.associateElement)==null||t.call(e,this))}disconnectedCallback(){var e,t;(t=(e=uc(this,pc))?.unassociateElement)==null||t.call(e,this),fc(this,pc,null)}attributeChangedCallback(e,t,n){var r,i,a,o;[x.MEDIA_PREVIEW_IMAGE,x.MEDIA_PREVIEW_COORDS].includes(e)&&this.update(),e===b.MEDIA_CONTROLLER&&(t&&((i=(r=uc(this,pc))?.unassociateElement)==null||i.call(r,this),fc(this,pc,null)),n&&this.isConnected&&(fc(this,pc,this.getRootNode()?.getElementById(n)),(o=(a=uc(this,pc))?.associateElement)==null||o.call(a,this)))}get mediaPreviewImage(){return F(this,x.MEDIA_PREVIEW_IMAGE)}set mediaPreviewImage(e){I(this,x.MEDIA_PREVIEW_IMAGE,e)}get mediaPreviewCoords(){let e=this.getAttribute(x.MEDIA_PREVIEW_COORDS);if(e)return e.split(/\s+/).map(e=>+e)}set mediaPreviewCoords(e){if(!e){this.removeAttribute(x.MEDIA_PREVIEW_COORDS);return}this.setAttribute(x.MEDIA_PREVIEW_COORDS,e.join(` `))}update(){let e=this.mediaPreviewCoords,t=this.mediaPreviewImage;if(!(e&&t))return;let[n,r,i,a]=e,o=t.split(`#`)[0],s=getComputedStyle(this),{maxWidth:c,maxHeight:l,minWidth:u,minHeight:d}=s,f=s.getPropertyValue(`--media-preview-thumbnail-object-fit`).trim()||`contain`,p,m;if(f===`fill`){let e=parseInt(c)/i,t=parseInt(l)/a,n=parseInt(u)/i,r=parseInt(d)/a;p=e<1?e:Math.max(e,n),m=t<1?t:Math.max(t,r)}else{let e=Math.min(parseInt(c)/i,parseInt(l)/a),t=Math.max(parseInt(u)/i,parseInt(d)/a),n=e<1?e:t>1?t:1;p=n,m=n}let{style:h}=A(this.shadowRoot,`:host`),g=A(this.shadowRoot,`img`).style,ee=this.shadowRoot.querySelector(`img`),te=Math.min(p,m)<1?`min`:`max`;h.setProperty(`${te}-width`,`initial`,`important`),h.setProperty(`${te}-height`,`initial`,`important`),h.width=`${i*p}px`,h.height=`${a*m}px`;let ne=()=>{g.width=`${this.imgWidth*p}px`,g.height=`${this.imgHeight*m}px`,g.display=`block`};ee.src!==o&&(ee.onload=()=>{this.imgWidth=ee.naturalWidth,this.imgHeight=ee.naturalHeight,ne(),ee.onload=null},ee.src=o,ne()),ne(),g.transform=`translate(-${n*p}px, -${r*m}px)`}};pc=new WeakMap,hc.shadowRootOptions={mode:`open`},hc.getTemplateHTML=mc,O.customElements.get(`media-preview-thumbnail`)||O.customElements.define(`media-preview-thumbnail`,hc);var gc=hc,_c=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},vc=(e,t,n)=>(_c(e,t,`read from private field`),n?n.call(e):t.get(e)),yc=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},bc=(e,t,n,r)=>(_c(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),xc,Sc=class extends ao{constructor(){super(),yc(this,xc,void 0),bc(this,xc,this.shadowRoot.querySelector(`slot`)),vc(this,xc).textContent=$e(0)}static get observedAttributes(){return[...super.observedAttributes,x.MEDIA_PREVIEW_TIME]}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===x.MEDIA_PREVIEW_TIME&&n!=null&&(vc(this,xc).textContent=$e(parseFloat(n)))}get mediaPreviewTime(){return j(this,x.MEDIA_PREVIEW_TIME)}set mediaPreviewTime(e){M(this,x.MEDIA_PREVIEW_TIME,e)}};xc=new WeakMap,O.customElements.get(`media-preview-time-display`)||O.customElements.define(`media-preview-time-display`,Sc);var Cc=Sc,wc={SEEK_OFFSET:`seekoffset`},Tc=30,Ec=e=>`
  <svg aria-hidden="true" viewBox="0 0 20 24">
    <defs>
      <style>.text{font-size:8px;font-family:Arial-BoldMT, Arial;font-weight:700;}</style>
    </defs>
    <text class="text value" transform="translate(2.18 19.87)">${e}</text>
    <path d="M10 6V3L4.37 7 10 10.94V8a5.54 5.54 0 0 1 1.9 10.48v2.12A7.5 7.5 0 0 0 10 6Z"/>
  </svg>`;function Dc(e,t){return`
    <slot name="icon">${Ec(t.seekOffset)}</slot>
  `}var Oc=(e,t)=>{e.setAttribute(`aria-label`,D(`seek back {seekOffset} seconds`,{seekOffset:t}))};function kc(){return D(`Seek backward`)}var Ac=0,jc=class extends K{static get observedAttributes(){return[...super.observedAttributes,x.MEDIA_CURRENT_TIME,wc.SEEK_OFFSET]}connectedCallback(){super.connectedCallback(),this.seekOffset=j(this,wc.SEEK_OFFSET,Tc)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),Oc(this,this.seekOffset),e===wc.SEEK_OFFSET&&(this.seekOffset=j(this,wc.SEEK_OFFSET,Tc))}get seekOffset(){return j(this,wc.SEEK_OFFSET,Tc)}set seekOffset(e){M(this,wc.SEEK_OFFSET,e),this.setAttribute(`aria-label`,D(`seek back {seekOffset} seconds`,{seekOffset:this.seekOffset})),gt(vt(this,`icon`),this.seekOffset)}get mediaCurrentTime(){return j(this,x.MEDIA_CURRENT_TIME,Ac)}set mediaCurrentTime(e){M(this,x.MEDIA_CURRENT_TIME,e)}handleClick(){let e=Math.max(this.mediaCurrentTime-this.seekOffset,0),t=new O.CustomEvent(y.MEDIA_SEEK_REQUEST,{composed:!0,bubbles:!0,detail:e});this.dispatchEvent(t)}};jc.getSlotTemplateHTML=Dc,jc.getTooltipContentHTML=kc,O.customElements.get(`media-seek-backward-button`)||O.customElements.define(`media-seek-backward-button`,jc);var Mc=jc,Nc={SEEK_OFFSET:`seekoffset`},Pc=30,Fc=e=>`
  <svg aria-hidden="true" viewBox="0 0 20 24">
    <defs>
      <style>.text{font-size:8px;font-family:Arial-BoldMT, Arial;font-weight:700;}</style>
    </defs>
    <text class="text value" transform="translate(8.9 19.87)">${e}</text>
    <path d="M10 6V3l5.61 4L10 10.94V8a5.54 5.54 0 0 0-1.9 10.48v2.12A7.5 7.5 0 0 1 10 6Z"/>
  </svg>`;function Ic(e,t){return`
    <slot name="icon">${Fc(t.seekOffset)}</slot>
  `}var Lc=(e,t)=>{e.setAttribute(`aria-label`,D(`seek forward {seekOffset} seconds`,{seekOffset:t}))};function Rc(){return D(`Seek forward`)}var zc=0,Bc=class extends K{static get observedAttributes(){return[...super.observedAttributes,x.MEDIA_CURRENT_TIME,Nc.SEEK_OFFSET]}connectedCallback(){super.connectedCallback(),this.seekOffset=j(this,Nc.SEEK_OFFSET,Pc)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),Lc(this,this.seekOffset),e===Nc.SEEK_OFFSET&&(this.seekOffset=j(this,Nc.SEEK_OFFSET,Pc))}get seekOffset(){return j(this,Nc.SEEK_OFFSET,Pc)}set seekOffset(e){M(this,Nc.SEEK_OFFSET,e),this.setAttribute(`aria-label`,D(`seek forward {seekOffset} seconds`,{seekOffset:this.seekOffset})),gt(vt(this,`icon`),this.seekOffset)}get mediaCurrentTime(){return j(this,x.MEDIA_CURRENT_TIME,zc)}set mediaCurrentTime(e){M(this,x.MEDIA_CURRENT_TIME,e)}handleClick(){let e=this.mediaCurrentTime+this.seekOffset,t=new O.CustomEvent(y.MEDIA_SEEK_REQUEST,{composed:!0,bubbles:!0,detail:e});this.dispatchEvent(t)}};Bc.getSlotTemplateHTML=Ic,Bc.getTooltipContentHTML=Rc,O.customElements.get(`media-seek-forward-button`)||O.customElements.define(`media-seek-forward-button`,Bc);var Vc=Bc,Hc=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Uc=(e,t,n)=>(Hc(e,t,`read from private field`),n?n.call(e):t.get(e)),Wc=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Gc=(e,t,n,r)=>(Hc(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),Kc=(e,t,n)=>(Hc(e,t,`access private method`),n),qc,Jc,Yc,Xc,Zc,Qc,$c,el,tl,nl,rl,il={REMAINING:`remaining`,SHOW_DURATION:`showduration`,NO_TOGGLE:`notoggle`},al=[...Object.values(il),x.MEDIA_CURRENT_TIME,x.MEDIA_DURATION,x.MEDIA_SEEKABLE],ol=[`Enter`,` `],sl=`&nbsp;/&nbsp;`,cl=(e,{timesSep:t=sl}={})=>{let n=e.mediaCurrentTime??0,[,r]=e.mediaSeekable??[],i=0;Number.isFinite(e.mediaDuration)?i=e.mediaDuration:Number.isFinite(r)&&(i=r);let a=e.remaining?$e(0-(i-n)):$e(n);return e.showDuration?`${a}${t}${$e(i)}`:a},ll=e=>{let t=e.mediaCurrentTime,[,n]=e.mediaSeekable??[],r=null;if(Number.isFinite(e.mediaDuration)?r=e.mediaDuration:Number.isFinite(n)&&(r=n),t==null||r===null){e.setAttribute(`aria-description`,D(`video not loaded, unknown time.`));return}let i=e.remaining?Qe(0-(r-t)):Qe(t);if(!e.showDuration){e.setAttribute(`aria-description`,i);return}let a=D(`{currentTime} of {totalTime}`,{currentTime:i,totalTime:Qe(r)});e.setAttribute(`aria-description`,a)};function ul(e,t){return`
    <slot>${cl(t)}</slot>
  `}var dl=e=>{e.setAttribute(`aria-label`,D(`playback time`))},fl=class extends ao{constructor(){super(),Wc(this,Xc),Wc(this,Qc),Wc(this,el),Wc(this,nl),Wc(this,qc,void 0),Wc(this,Jc,null),Wc(this,Yc,e=>{let{metaKey:t,altKey:n,key:r}=e;if(t||n||!ol.includes(r)){this.removeEventListener(`keyup`,Uc(this,Jc));return}this.addEventListener(`keyup`,Uc(this,Jc))}),Gc(this,qc,this.shadowRoot.querySelector(`slot`)),Uc(this,qc).innerHTML=`${cl(this)}`}static get observedAttributes(){return[...super.observedAttributes,...al,`disabled`]}connectedCallback(){let{style:e}=A(this.shadowRoot,`:host(:hover:not([notoggle]))`);e.setProperty(`cursor`,`var(--media-cursor, pointer)`),e.setProperty(`background`,`var(--media-control-hover-background, rgba(50 50 70 / .7))`),this.setAttribute(`aria-label`,D(`playback time`)),Kc(this,el,tl).call(this),super.connectedCallback()}toggleTimeDisplay(){this.noToggle||(this.hasAttribute(`remaining`)?this.removeAttribute(`remaining`):this.setAttribute(`remaining`,``))}disconnectedCallback(){this.disable(),Kc(this,Qc,$c).call(this),super.disconnectedCallback()}attributeChangedCallback(e,t,n){dl(this),al.includes(e)?this.update():e===`disabled`&&n!==t?n==null?Kc(this,el,tl).call(this):Kc(this,nl,rl).call(this):e===il.NO_TOGGLE&&n!==t&&(this.noToggle?Kc(this,nl,rl).call(this):Kc(this,el,tl).call(this)),super.attributeChangedCallback(e,t,n)}enable(){this.noToggle||(this.tabIndex=0)}disable(){this.tabIndex=-1}get remaining(){return N(this,il.REMAINING)}set remaining(e){P(this,il.REMAINING,e)}get showDuration(){return N(this,il.SHOW_DURATION)}set showDuration(e){P(this,il.SHOW_DURATION,e)}get noToggle(){return N(this,il.NO_TOGGLE)}set noToggle(e){P(this,il.NO_TOGGLE,e)}get mediaDuration(){return j(this,x.MEDIA_DURATION)}set mediaDuration(e){M(this,x.MEDIA_DURATION,e)}get mediaCurrentTime(){return j(this,x.MEDIA_CURRENT_TIME)}set mediaCurrentTime(e){M(this,x.MEDIA_CURRENT_TIME,e)}get mediaSeekable(){let e=this.getAttribute(x.MEDIA_SEEKABLE);if(e)return e.split(`:`).map(e=>+e)}set mediaSeekable(e){if(e==null){this.removeAttribute(x.MEDIA_SEEKABLE);return}this.setAttribute(x.MEDIA_SEEKABLE,e.join(`:`))}update(){let e=cl(this);ll(this),e!==Uc(this,qc).innerHTML&&(Uc(this,qc).innerHTML=e)}};qc=new WeakMap,Jc=new WeakMap,Yc=new WeakMap,Xc=new WeakSet,Zc=function(){Uc(this,Jc)||(Gc(this,Jc,e=>{let{key:t}=e;if(!ol.includes(t)){this.removeEventListener(`keyup`,Uc(this,Jc));return}this.toggleTimeDisplay()}),this.addEventListener(`keydown`,Uc(this,Yc)),this.addEventListener(`click`,this.toggleTimeDisplay))},Qc=new WeakSet,$c=function(){Uc(this,Jc)&&(this.removeEventListener(`keyup`,Uc(this,Jc)),this.removeEventListener(`keydown`,Uc(this,Yc)),this.removeEventListener(`click`,this.toggleTimeDisplay),Gc(this,Jc,null))},el=new WeakSet,tl=function(){!this.noToggle&&!this.hasAttribute(`disabled`)&&(this.setAttribute(`role`,`button`),this.enable(),Kc(this,Xc,Zc).call(this))},nl=new WeakSet,rl=function(){this.removeAttribute(`role`),this.disable(),Kc(this,Qc,$c).call(this)},fl.getSlotTemplateHTML=ul,O.customElements.get(`media-time-display`)||O.customElements.define(`media-time-display`,fl);var pl=fl,ml=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},Y=(e,t,n)=>(ml(e,t,`read from private field`),n?n.call(e):t.get(e)),hl=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},gl=(e,t,n,r)=>(ml(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),_l=(e,t,n,r)=>({set _(r){gl(e,t,r,n)},get _(){return Y(e,t,r)}}),vl,yl,bl,xl,Sl,Cl,wl,Tl,El,Dl,Ol=class{constructor(e,t,n){hl(this,vl,void 0),hl(this,yl,void 0),hl(this,bl,void 0),hl(this,xl,void 0),hl(this,Sl,void 0),hl(this,Cl,void 0),hl(this,wl,void 0),hl(this,Tl,void 0),hl(this,El,0),hl(this,Dl,(e=performance.now())=>{gl(this,El,requestAnimationFrame(Y(this,Dl))),gl(this,xl,performance.now()-Y(this,bl));let t=1e3/this.fps;if(Y(this,xl)>t){gl(this,bl,e-Y(this,xl)%t);let n=1e3/((e-Y(this,yl))/++_l(this,Sl)._),r=(e-Y(this,Cl))/1e3/this.duration,i=Y(this,wl)+r*this.playbackRate;i-Y(this,vl).valueAsNumber>0?gl(this,Tl,this.playbackRate/this.duration/n):(gl(this,Tl,.995*Y(this,Tl)),i=Y(this,vl).valueAsNumber+Y(this,Tl)),this.callback(i)}}),gl(this,vl,e),this.callback=t,this.fps=n}start(){Y(this,El)===0&&(gl(this,bl,performance.now()),gl(this,yl,Y(this,bl)),gl(this,Sl,0),Y(this,Dl).call(this))}stop(){Y(this,El)!==0&&(cancelAnimationFrame(Y(this,El)),gl(this,El,0))}update({start:e,duration:t,playbackRate:n}){let r=e-Y(this,vl).valueAsNumber,i=Math.abs(t-this.duration);(r>0||r<-.03||i>=.5)&&this.callback(e),gl(this,wl,e),gl(this,Cl,performance.now()),this.duration=t,this.playbackRate=n}};vl=new WeakMap,yl=new WeakMap,bl=new WeakMap,xl=new WeakMap,Sl=new WeakMap,Cl=new WeakMap,wl=new WeakMap,Tl=new WeakMap,El=new WeakMap,Dl=new WeakMap;var kl=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},X=(e,t,n)=>(kl(e,t,`read from private field`),n?n.call(e):t.get(e)),Z=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},Al=(e,t,n,r)=>(kl(e,t,`write to private field`),r?r.call(e,n):t.set(e,n),n),jl=(e,t,n)=>(kl(e,t,`access private method`),n),Ml,Nl,Pl,Fl,Il,Ll,Rl,zl,Bl,Vl,Hl,Ul,Wl,Gl,Kl,ql,Jl,Yl,Xl,Zl,Ql,$l,eu,tu,nu,ru,iu=e=>{let t=e.range,n=Qe(+su(e)),r=Qe(+e.mediaSeekableEnd),i=n&&r?D(`{currentTime} of {totalTime}`,{currentTime:n,totalTime:r}):D(`video not loaded, unknown time.`);t.setAttribute(`aria-valuetext`,i)};function au(e){return`
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

      :host(:is([${x.MEDIA_PREVIEW_IMAGE}], [${x.MEDIA_PREVIEW_TIME}])[dragging]) [part~="preview-box"] {
        transition-duration: var(--media-preview-transition-duration-in, .5s);
        transition-delay: var(--media-preview-transition-delay-in, .25s);
        visibility: visible;
        opacity: 1;
      }

      @media (hover: hover) {
        :host(:is([${x.MEDIA_PREVIEW_IMAGE}], [${x.MEDIA_PREVIEW_TIME}]):hover) [part~="preview-box"] {
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

      :host([${x.MEDIA_PREVIEW_IMAGE}][dragging]) media-preview-thumbnail,
      :host([${x.MEDIA_PREVIEW_IMAGE}][dragging]) ::slotted(media-preview-thumbnail) {
        transition-delay: var(--media-preview-transition-delay-in, .25s);
        visibility: visible;
      }

      @media (hover: hover) {
        :host([${x.MEDIA_PREVIEW_IMAGE}]:hover) media-preview-thumbnail,
        :host([${x.MEDIA_PREVIEW_IMAGE}]:hover) ::slotted(media-preview-thumbnail) {
          transition-delay: var(--media-preview-transition-delay-in, .25s);
          visibility: visible;
        }

        :host([${x.MEDIA_PREVIEW_TIME}]:hover) {
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

      :host([${x.MEDIA_PREVIEW_IMAGE}]) media-preview-chapter-display,
      :host([${x.MEDIA_PREVIEW_IMAGE}]) ::slotted(media-preview-chapter-display) {
        transition-delay: var(--media-preview-transition-delay-in, .25s);
        border-radius: var(--media-preview-chapter-border-radius, 0);
        padding: var(--media-preview-chapter-padding, 3.5px 9px 0);
        margin: var(--media-preview-chapter-margin, 0);
        min-width: 100%;
      }

      media-preview-chapter-display[${x.MEDIA_PREVIEW_CHAPTER}],
      ::slotted(media-preview-chapter-display[${x.MEDIA_PREVIEW_CHAPTER}]) {
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

      :host([${x.MEDIA_PREVIEW_IMAGE}]) media-preview-time-display,
      :host([${x.MEDIA_PREVIEW_IMAGE}]) ::slotted(media-preview-time-display) {
        transition-delay: var(--media-preview-transition-delay-in, .25s);
        border-radius: var(--media-preview-time-border-radius,
          0 0 var(--media-preview-border-radius) var(--media-preview-border-radius));
        min-width: 100%;
      }

      :host([${x.MEDIA_PREVIEW_TIME}]:hover) {
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
          <template shadowrootmode="${gc.shadowRootOptions.mode}">
            ${gc.getTemplateHTML({})}
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
  `}var ou=(e,t=e.mediaCurrentTime)=>{let n=Number.isFinite(e.mediaSeekableStart)?e.mediaSeekableStart:0,r=Number.isFinite(e.mediaDuration)?e.mediaDuration:e.mediaSeekableEnd;if(Number.isNaN(r))return 0;let i=(t-n)/(r-n);return Math.max(0,Math.min(i,1))},su=(e,t=e.range.valueAsNumber)=>{let n=Number.isFinite(e.mediaSeekableStart)?e.mediaSeekableStart:0,r=Number.isFinite(e.mediaDuration)?e.mediaDuration:e.mediaSeekableEnd;return Number.isNaN(r)?0:t*(r-n)+n},cu=class extends Ha{constructor(){super(),Z(this,Ul),Z(this,Kl),Z(this,Jl),Z(this,Xl),Z(this,Ql),Z(this,eu),Z(this,nu),Z(this,Ml,null),Z(this,Nl,void 0),Z(this,Pl,void 0),Z(this,Fl,void 0),Z(this,Il,void 0),Z(this,Ll,void 0),Z(this,Rl,void 0),Z(this,zl,void 0),Z(this,Bl,void 0),Z(this,Vl,void 0),Z(this,Hl,()=>{jl(this,Ul,Wl).call(this)?X(this,Nl).start():X(this,Nl).stop()}),Z(this,Gl,e=>{this.dragging||(Ue(e)&&(this.range.valueAsNumber=e),X(this,Vl)||this.updateBar())}),this.shadowRoot.querySelector(`#track`).insertAdjacentHTML(`afterbegin`,`<div id="buffered" part="buffered"></div>`),Al(this,Pl,this.shadowRoot.querySelectorAll(`[part~="box"]`)),Al(this,Il,this.shadowRoot.querySelector(`[part~="preview-box"]`)),Al(this,Ll,this.shadowRoot.querySelector(`[part~="current-box"]`));let e=getComputedStyle(this);Al(this,Rl,parseInt(e.getPropertyValue(`--media-box-padding-left`))),Al(this,zl,parseInt(e.getPropertyValue(`--media-box-padding-right`))),Al(this,Nl,new Ol(this.range,X(this,Gl),60))}static get observedAttributes(){return[...super.observedAttributes,x.MEDIA_PAUSED,x.MEDIA_DURATION,x.MEDIA_SEEKABLE,x.MEDIA_CURRENT_TIME,x.MEDIA_PREVIEW_IMAGE,x.MEDIA_PREVIEW_TIME,x.MEDIA_PREVIEW_CHAPTER,x.MEDIA_BUFFERED,x.MEDIA_PLAYBACK_RATE,x.MEDIA_LOADING,x.MEDIA_ENDED]}connectedCallback(){var e;super.connectedCallback(),this.range.setAttribute(`aria-label`,D(`seek`)),X(this,Hl).call(this),Al(this,Ml,this.getRootNode()),(e=X(this,Ml))==null||e.addEventListener(`transitionstart`,this)}disconnectedCallback(){var e;super.disconnectedCallback(),X(this,Nl).stop(),(e=X(this,Ml))==null||e.removeEventListener(`transitionstart`,this),Al(this,Ml,null)}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),t!=n&&(e===x.MEDIA_CURRENT_TIME||e===x.MEDIA_PAUSED||e===x.MEDIA_ENDED||e===x.MEDIA_LOADING||e===x.MEDIA_DURATION||e===x.MEDIA_SEEKABLE?(X(this,Nl).update({start:ou(this),duration:this.mediaSeekableEnd-this.mediaSeekableStart,playbackRate:this.mediaPlaybackRate}),X(this,Hl).call(this),iu(this)):e===x.MEDIA_BUFFERED&&this.updateBufferedBar(),(e===x.MEDIA_DURATION||e===x.MEDIA_SEEKABLE)&&(this.mediaChaptersCues=X(this,Bl),this.updateBar()))}get mediaChaptersCues(){return X(this,Bl)}set mediaChaptersCues(e){Al(this,Bl,e),this.updateSegments(X(this,Bl)?.map(e=>({start:ou(this,e.startTime),end:ou(this,e.endTime)})))}get mediaPaused(){return N(this,x.MEDIA_PAUSED)}set mediaPaused(e){P(this,x.MEDIA_PAUSED,e)}get mediaLoading(){return N(this,x.MEDIA_LOADING)}set mediaLoading(e){P(this,x.MEDIA_LOADING,e)}get mediaDuration(){return j(this,x.MEDIA_DURATION)}set mediaDuration(e){M(this,x.MEDIA_DURATION,e)}get mediaCurrentTime(){return j(this,x.MEDIA_CURRENT_TIME)}set mediaCurrentTime(e){M(this,x.MEDIA_CURRENT_TIME,e)}get mediaPlaybackRate(){return j(this,x.MEDIA_PLAYBACK_RATE,1)}set mediaPlaybackRate(e){M(this,x.MEDIA_PLAYBACK_RATE,e)}get mediaBuffered(){let e=this.getAttribute(x.MEDIA_BUFFERED);return e?e.split(` `).map(e=>e.split(`:`).map(e=>+e)):[]}set mediaBuffered(e){if(!e){this.removeAttribute(x.MEDIA_BUFFERED);return}let t=e.map(e=>e.join(`:`)).join(` `);this.setAttribute(x.MEDIA_BUFFERED,t)}get mediaSeekable(){let e=this.getAttribute(x.MEDIA_SEEKABLE);if(e)return e.split(`:`).map(e=>+e)}set mediaSeekable(e){if(e==null){this.removeAttribute(x.MEDIA_SEEKABLE);return}this.setAttribute(x.MEDIA_SEEKABLE,e.join(`:`))}get mediaSeekableEnd(){let[,e=this.mediaDuration]=this.mediaSeekable??[];return e}get mediaSeekableStart(){let[e=0]=this.mediaSeekable??[];return e}get mediaPreviewImage(){return F(this,x.MEDIA_PREVIEW_IMAGE)}set mediaPreviewImage(e){I(this,x.MEDIA_PREVIEW_IMAGE,e)}get mediaPreviewTime(){return j(this,x.MEDIA_PREVIEW_TIME)}set mediaPreviewTime(e){M(this,x.MEDIA_PREVIEW_TIME,e)}get mediaEnded(){return N(this,x.MEDIA_ENDED)}set mediaEnded(e){P(this,x.MEDIA_ENDED,e)}updateBar(){super.updateBar(),this.updateBufferedBar(),this.updateCurrentBox()}updateBufferedBar(){let e=this.mediaBuffered;if(!e.length)return;let t;if(this.mediaEnded)t=1;else{let n=this.mediaCurrentTime,[,r=this.mediaSeekableStart]=e.find(([e,t])=>e<=n&&n<=t)??[];t=ou(this,r)}let{style:n}=A(this.shadowRoot,`#buffered`);n.setProperty(`width`,`${t*100}%`)}updateCurrentBox(){if(!this.shadowRoot.querySelector(`slot[name="current"]`).assignedElements().length)return;let e=A(this.shadowRoot,`#current-rail`),t=A(this.shadowRoot,`[part~="current-box"]`),n=jl(this,Kl,ql).call(this,X(this,Ll)),r=jl(this,Jl,Yl).call(this,n,this.range.valueAsNumber),i=jl(this,Xl,Zl).call(this,n,this.range.valueAsNumber);e.style.transform=`translateX(${r})`,e.style.setProperty(`--_range-width`,`${n.range.width}`),t.style.setProperty(`--_box-shift`,`${i}`),t.style.setProperty(`--_box-width`,`${n.box.width}px`),t.style.setProperty(`visibility`,`initial`)}handleEvent(e){switch(super.handleEvent(e),e.type){case`input`:jl(this,nu,ru).call(this);break;case`pointermove`:jl(this,Ql,$l).call(this,e);break;case`pointerup`:X(this,Vl)&&Al(this,Vl,!1);break;case`pointerdown`:Al(this,Vl,!0);break;case`pointerleave`:jl(this,eu,tu).call(this,null);break;case`transitionstart`:yt(e.target,this)&&setTimeout(()=>X(this,Hl).call(this),0);break}}};Ml=new WeakMap,Nl=new WeakMap,Pl=new WeakMap,Fl=new WeakMap,Il=new WeakMap,Ll=new WeakMap,Rl=new WeakMap,zl=new WeakMap,Bl=new WeakMap,Vl=new WeakMap,Hl=new WeakMap,Ul=new WeakSet,Wl=function(){return this.isConnected&&!this.mediaPaused&&!this.mediaLoading&&!this.mediaEnded&&this.mediaSeekableEnd>0&&Ct(this)},Gl=new WeakMap,Kl=new WeakSet,ql=function(e){let t=((this.getAttribute(`bounds`)?bt(this,`#${this.getAttribute(`bounds`)}`):this.parentElement)??this).getBoundingClientRect(),n=this.range.getBoundingClientRect(),r=e.offsetWidth;return{box:{width:r,min:-(n.left-t.left-r/2),max:t.right-n.left-r/2},bounds:t,range:n}},Jl=new WeakSet,Yl=function(e,t){let n=`${t*100}%`,{width:r,min:i,max:a}=e.box;if(!r)return n;if(Number.isNaN(i)||(n=`max(${`calc(1 / var(--_range-width) * 100 * ${i}% + var(--media-box-padding-left))`}, ${n})`),!Number.isNaN(a)){let e=`calc(1 / var(--_range-width) * 100 * ${a}% - var(--media-box-padding-right))`;n=`min(${n}, ${e})`}return n},Xl=new WeakSet,Zl=function(e,t){let{width:n,min:r,max:i}=e.box,a=t*e.range.width;if(a<r+X(this,Rl)){let t=e.range.left-e.bounds.left-X(this,Rl);return`${a-n/2+t}px`}if(a>i-X(this,zl)){let t=e.bounds.right-e.range.right-X(this,zl);return`${a+n/2-t-e.range.width}px`}return 0},Ql=new WeakSet,$l=function(e){let t=[...X(this,Pl)].some(t=>e.composedPath().includes(t));if(!this.dragging&&(t||!e.composedPath().includes(this))){jl(this,eu,tu).call(this,null);return}let n=this.mediaSeekableEnd;if(!n)return;let r=A(this.shadowRoot,`#preview-rail`),i=A(this.shadowRoot,`[part~="preview-box"]`),a=jl(this,Kl,ql).call(this,X(this,Il)),o=(e.clientX-a.range.left)/a.range.width;o=Math.max(0,Math.min(1,o));let s=jl(this,Jl,Yl).call(this,a,o),c=jl(this,Xl,Zl).call(this,a,o);r.style.transform=`translateX(${s})`,r.style.setProperty(`--_range-width`,`${a.range.width}`),i.style.setProperty(`--_box-shift`,`${c}`),i.style.setProperty(`--_box-width`,`${a.box.width}px`);let l=Math.round(X(this,Fl))-Math.round(o*n);Math.abs(l)<1&&o>.01&&o<.99||(Al(this,Fl,o*n),jl(this,eu,tu).call(this,X(this,Fl)))},eu=new WeakSet,tu=function(e){this.dispatchEvent(new O.CustomEvent(y.MEDIA_PREVIEW_REQUEST,{composed:!0,bubbles:!0,detail:e}))},nu=new WeakSet,ru=function(){X(this,Nl).stop();let e=su(this);this.dispatchEvent(new O.CustomEvent(y.MEDIA_SEEK_REQUEST,{composed:!0,bubbles:!0,detail:e}))},cu.shadowRootOptions={mode:`open`},cu.getContainerTemplateHTML=au,O.customElements.get(`media-time-range`)||O.customElements.define(`media-time-range`,cu);var lu=cu,uu=(e,t,n)=>{if(!t.has(e))throw TypeError(`Cannot `+n)},du=(e,t,n)=>(uu(e,t,`read from private field`),n?n.call(e):t.get(e)),fu=(e,t,n)=>{if(t.has(e))throw TypeError(`Cannot add the same private member more than once`);t instanceof WeakSet?t.add(e):t.set(e,n)},pu,mu=1,hu=e=>e.mediaMuted?0:e.mediaVolume,gu=e=>`${Math.round(e*100)}%`,_u=class extends Ha{constructor(){super(...arguments),fu(this,pu,()=>{let e=this.range.value,t=new O.CustomEvent(y.MEDIA_VOLUME_REQUEST,{composed:!0,bubbles:!0,detail:e});this.dispatchEvent(t)})}static get observedAttributes(){return[...super.observedAttributes,x.MEDIA_VOLUME,x.MEDIA_MUTED,x.MEDIA_VOLUME_UNAVAILABLE]}connectedCallback(){super.connectedCallback(),this.range.setAttribute(`aria-label`,D(`volume`)),this.range.addEventListener(`input`,du(this,pu))}disconnectedCallback(){this.range.removeEventListener(`input`,du(this,pu)),super.disconnectedCallback()}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),(e===x.MEDIA_VOLUME||e===x.MEDIA_MUTED)&&(this.range.valueAsNumber=hu(this),this.range.setAttribute(`aria-valuetext`,gu(this.range.valueAsNumber)),this.updateBar())}get mediaVolume(){return j(this,x.MEDIA_VOLUME,mu)}set mediaVolume(e){M(this,x.MEDIA_VOLUME,e)}get mediaMuted(){return N(this,x.MEDIA_MUTED)}set mediaMuted(e){P(this,x.MEDIA_MUTED,e)}get mediaVolumeUnavailable(){return F(this,x.MEDIA_VOLUME_UNAVAILABLE)}set mediaVolumeUnavailable(e){I(this,x.MEDIA_VOLUME_UNAVAILABLE,e)}};pu=new WeakMap,O.customElements.get(`media-volume-range`)||O.customElements.define(`media-volume-range`,_u);var vu=_u;function yu(e){return`
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

        :host([${x.MEDIA_LOOP}]) #checked-indicator {
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
    `}function bu(){return D(`Loop`)}var xu=class extends K{constructor(){super(...arguments),this.container=null}static get observedAttributes(){return[...super.observedAttributes,x.MEDIA_LOOP]}connectedCallback(){super.connectedCallback(),this.container=this.shadowRoot?.querySelector(`#icon`)||null,this.container&&(this.container.textContent=D(`Loop`))}attributeChangedCallback(e,t,n){super.attributeChangedCallback(e,t,n),e===x.MEDIA_LOOP&&this.container&&this.setAttribute(`aria-checked`,this.mediaLoop?`true`:`false`)}get mediaLoop(){return N(this,x.MEDIA_LOOP)}set mediaLoop(e){P(this,x.MEDIA_LOOP,e)}handleClick(){let e=!this.mediaLoop,t=new O.CustomEvent(y.MEDIA_LOOP_REQUEST,{composed:!0,bubbles:!0,detail:e});this.dispatchEvent(t)}};xu.getSlotTemplateHTML=yu,xu.getTooltipContentHTML=bu,O.customElements.get(`media-loop-button`)||O.customElements.define(`media-loop-button`,xu);var Su=xu;function Q(e){if(typeof e==`boolean`)return e?``:void 0;if(typeof e!=`function`){if(Array.isArray(e)&&e.every(e=>typeof e==`string`||typeof e==`number`||typeof e==`boolean`))return e.join(` `);if(!(typeof e==`object`&&e))return e}}v({tagName:`media-gesture-receiver`,elementClass:Pt,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),v({tagName:`media-container`,elementClass:pn,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var Cu=v({tagName:`media-controller`,elementClass:Qr,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});v({tagName:`media-tooltip`,elementClass:ni,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),v({tagName:`media-chrome-button`,elementClass:vi,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),v({tagName:`media-airplay-button`,elementClass:wi,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var wu=v({tagName:`media-captions-button`,elementClass:Ni,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});v({tagName:`media-cast-button`,elementClass:Bi,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),v({tagName:`media-chrome-dialog`,elementClass:da,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),v({tagName:`media-chrome-range`,elementClass:Ua,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var Tu=v({tagName:`media-control-bar`,elementClass:Za,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});v({tagName:`media-text-display`,elementClass:oo,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),v({tagName:`media-duration-display`,elementClass:ho,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),v({tagName:`media-error-dialog`,elementClass:ko,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),v({tagName:`media-keyboard-shortcuts-dialog`,elementClass:Ro,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var Eu=v({tagName:`media-fullscreen-button`,elementClass:Xo,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});v({tagName:`media-live-button`,elementClass:as,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),v({tagName:`media-loading-indicator`,elementClass:_s,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var Du=v({tagName:`media-mute-button`,elementClass:Ts,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Ou=v({tagName:`media-pip-button`,elementClass:js,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),ku=v({tagName:`media-playback-rate-button`,elementClass:Hs,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Au=v({tagName:`media-play-button`,elementClass:Ys,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});v({tagName:`media-poster-image`,elementClass:tc,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),v({tagName:`media-preview-chapter-display`,elementClass:cc,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),v({tagName:`media-preview-thumbnail`,elementClass:gc,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),v({tagName:`media-preview-time-display`,elementClass:Cc,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var ju=v({tagName:`media-seek-backward-button`,elementClass:Mc,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Mu=v({tagName:`media-seek-forward-button`,elementClass:Vc,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Nu=v({tagName:`media-time-display`,elementClass:pl,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Pu=v({tagName:`media-time-range`,elementClass:lu,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}}),Fu=v({tagName:`media-volume-range`,elementClass:vu,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});v({tagName:`media-loop-button`,elementClass:Su,react:_.default,toAttributeValue:Q,defaultProps:{suppressHydrationWarning:!0}});var $=a(),Iu={noTooltip:!0};function Lu({isMobile:e,showPip:t,showCaptionsButton:n,topChrome:r,extraControls:i,children:a,className:o,style:s}){return(0,$.jsxs)(Cu,{className:`ts-player ${o||``}`.trim(),autohide:`2`,style:s,children:[a,(0,$.jsx)(`div`,{slot:`top-chrome`,className:`ts-player-top`,children:r}),(0,$.jsxs)(Tu,{className:`ts-player-bar`,children:[(0,$.jsx)(Au,{...Iu,className:e?`ts-player-play-lg`:void 0}),(0,$.jsx)(ju,{...Iu,seekOffset:10}),(0,$.jsx)(Mu,{...Iu,seekOffset:10}),(0,$.jsx)(Nu,{showDuration:!0}),(0,$.jsx)(Pu,{}),(0,$.jsx)(Du,{...Iu}),e?null:(0,$.jsx)(Fu,{}),i,n?(0,$.jsx)(wu,{...Iu}):null,(0,$.jsx)(ku,{...Iu,rates:[.75,1,1.25,1.5,2]}),t?(0,$.jsx)(Ou,{...Iu}):null,(0,$.jsx)(Eu,{...Iu})]})]})}var Ru=e=>!!(e.canPlayType(`application/vnd.apple.mpegurl`)||e.canPlayType(`application/x-mpegURL`));function zu({enabled:e,open:t,videoRef:n,videoEpoch:r,src:i,playbackRate:a=1,onLoading:o,onError:s,onNotSupported:c,onSubtitleTracks:l,onSubtitleTrack:u,onDuration:d}){let f=(0,_.useRef)(null),h=(0,_.useRef)(o),g=(0,_.useRef)(s),ee=(0,_.useRef)(c),te=(0,_.useRef)(l),ne=(0,_.useRef)(u),re=(0,_.useRef)(d);return(0,_.useEffect)(()=>{h.current=o,g.current=s,ee.current=c,te.current=l,ne.current=u,re.current=d}),(0,_.useEffect)(()=>{let r=n.current;if(!e||!t||!r||!i)return;let o=null,s=!1;h.current?.(!0),te.current?.([]);let c=()=>{let e=(o?.subtitleTracks||[]).map((e,t)=>({id:t,name:e.name,lang:e.lang}));te.current?.(e)};if(Te.isSupported()){let e=p();o=new Te({xhrSetup:t=>{e&&t.setRequestHeader(`Authorization`,e)}}),f.current=o,o.on(Te.Events.MANIFEST_PARSED,()=>{c(),ne.current?.(o?.subtitleTrack??-1),r.playbackRate=a,r.play().catch(()=>{})}),o.on(Te.Events.LEVEL_LOADED,(e,t)=>{let n=t.details?.totalduration;typeof n==`number`&&re.current?.(n)}),o.on(Te.Events.SUBTITLE_TRACKS_UPDATED,c),o.on(Te.Events.SUBTITLE_TRACK_SWITCH,(e,t)=>ne.current?.(t.id)),o.on(Te.Events.ERROR,(e,t)=>{!t.fatal||!o||(t.type===Te.ErrorTypes.NETWORK_ERROR?o.startLoad():t.type===Te.ErrorTypes.MEDIA_ERROR?o.recoverMediaError():(o.stopLoad(),h.current?.(!1),g.current?.()))}),o.loadSource(i),o.attachMedia(r)}else Ru(r)?(s=!0,r.src=m(i),r.load(),r.playbackRate=a,r.play().catch(()=>{})):ee.current?.();return()=>{f.current===o&&(f.current=null),o?.destroy(),s&&(r.pause(),r.removeAttribute(`src`),r.load()),te.current?.([]),ne.current?.(-1)}},[e,t,n,r,i]),f}var Bu=5e3,Vu=5;function Hu({enabled:e,hash:t,fileIndex:n,title:r,initialTimecode:i=0,videoRef:a,onViewedChange:o}){let s=(0,_.useRef)(0),c=(0,_.useRef)(!1),l=(0,_.useRef)(o);(0,_.useEffect)(()=>{l.current=o},[o]),(0,_.useEffect)(()=>{c.current=!1},[t,n,i]);let u=(0,_.useCallback)(async i=>{if(!(!e||!t||n==null))try{await ve(t,n,i),ye({hash:t,fileIndex:n,title:r||t,fileName:r||String(n),timecode:i}),l.current?.()}catch{}},[e,t,n,r]);return{flushTimecode:(0,_.useCallback)(()=>{let t=a.current;!t||!e||u(t.currentTime)},[a,e,u]),onTimeUpdate:(0,_.useCallback)(()=>{let t=a.current;if(!t||!e)return;let n=Date.now();n-s.current<Bu||(s.current=n,u(t.currentTime))},[a,e,u]),applyResumeIfNeeded:(0,_.useCallback)(()=>{let e=a.current;if(!e||c.current)return;if(!(i>Vu)){c.current=!0;return}let t=e.duration;if(Number.isFinite(t)&&t>0&&i>=t-Vu){c.current=!0;return}e.currentTime=i,c.current=!0},[a,i]),saveTimecode:u}}var Uu=3e4;function Wu({tip:e,children:t}){return(0,$.jsx)(`span`,{className:`inline-flex min-w-0 max-w-full items-center gap-1`,title:e,children:t})}function Gu(e){switch(e.split(`?`)[0].split(`.`).pop()?.toLowerCase()){case`mp4`:return`video/mp4`;case`ogg`:case`ogv`:return`video/ogg`;case`webm`:return`video/webm`;default:return``}}var Ku=e=>!!(e.canPlayType(`application/vnd.apple.mpegurl`)||e.canPlayType(`application/x-mpegURL`)),qu=()=>typeof document<`u`&&`pictureInPictureEnabled`in document&&!!document.pictureInPictureEnabled;function Ju({videoSrc:e,downloadSrc:n=e,title:a,onNotSupported:p,hls:g=!1,heartbeatSrc:ve=``,showTrigger:ye=!0,inlineTrigger:De=!1,inlineTriggerPrimary:ke=!1,initiallyOpen:Ae=!1,onClose:je,captionSrc:Me,hash:v,fileIndex:Ne,initialTimecode:y=0,trackTimecode:b=!1,onViewedChange:Pe,audioTracks:Fe=[],audioIndex:x=0,onAudioIndexChange:Ie}){let{t:S}=ne(),C=i(_e(`dialog`)),{setImmersive:Le}=d(),Re=(0,_.useRef)(null),w=(0,_.useRef)(p),T=(0,_.useRef)(null),[E,ze]=(0,_.useState)(Ae),[Be,Ve]=(0,_.useState)(0),[He,Ue]=(0,_.useState)(!0),[We,Ge]=(0,_.useState)(!1),[Ke,qe]=(0,_.useState)(!1),[Je,Ye]=(0,_.useState)(!1),[D,Xe]=(0,_.useState)([]),[Ze,Qe]=(0,_.useState)(!1),[$e,et]=(0,_.useState)(x),[tt,nt]=(0,_.useState)(e),[rt,it]=(0,_.useState)(Fe),{data:at}=we(E?v:void 0),ot=Ce(v,{enabled:E&&!!v,fast:!1}),st=!!(v&&Ne!=null&&b),O=g&&!!v&&Ne!=null&&rt.length>1,k=qu(),ct=g?D.length>0:!!Me,lt=C?Se:Je?xe:be,ut=C?void 0:Je?`min(86dvh, calc(100dvh - 4rem))`:`min(72dvh, 40rem)`,dt=(0,_.useCallback)(e=>{Re.current=e,Ve(e=>e+1)},[]);(0,_.useEffect)(()=>{w.current=p},[p]),(0,_.useEffect)(()=>{nt(e),et(x),it(Fe),Ge(!1),Ue(!0)},[e,x,Fe]),zu({enabled:g,open:E,videoRef:Re,videoEpoch:Be,src:tt,onLoading:Ue,onError:()=>Ge(!0),onNotSupported:()=>w.current?.(),onSubtitleTracks:Xe});let{flushTimecode:ft,onTimeUpdate:pt,applyResumeIfNeeded:mt}=Hu({enabled:st,hash:v,fileIndex:Ne,title:a,initialTimecode:y,videoRef:Re,onViewedChange:Pe}),ht=(0,_.useCallback)(()=>ze(!0),[]),gt=(0,_.useCallback)(()=>{ft(),ze(!1),Ge(!1),Qe(!1),Ye(!1),je?.()},[ft,je]),_t=c({isOpen:E,onOpenChange:e=>{e?ze(!0):gt()}});f(E),(0,_.useEffect)(()=>{if(!C||!E){Le(!1);return}return Le(!0),()=>Le(!1)},[C,E,Le]),(0,_.useEffect)(()=>{if(!E){Oe(null);return}return Oe(null),()=>Oe(null)},[E]),(0,_.useEffect)(()=>{let e=document.createElement(`video`);(g?Te.isSupported()||Ku(e):e.canPlayType(Gu(tt)))||w.current?.()},[g,tt]),(0,_.useEffect)(()=>{let e=Re.current;if(!(!E||g||!e))return Ue(!0),e.src=m(tt),e.load(),e.play().catch(()=>{}),()=>{e.pause(),e.removeAttribute(`src`),e.load()}},[E,g,Be,tt]),(0,_.useEffect)(()=>{if(!E||!ve)return;let e=window.setInterval(()=>{h(ve,{cache:`no-store`}).catch(()=>{})},Uu);return()=>window.clearInterval(e)},[ve,E]),(0,_.useEffect)(()=>{let e=()=>qe(!!document.fullscreenElement);return document.addEventListener(`fullscreenchange`,e),()=>document.removeEventListener(`fullscreenchange`,e)},[]),(0,_.useEffect)(()=>{let e=Re.current;if(!e||T.current==null)return;let t=T.current,n=()=>{T.current!=null&&(e.currentTime=t,T.current=null,e.play().catch(()=>{}))};return e.addEventListener(`loadedmetadata`,n,{once:!0}),()=>e.removeEventListener(`loadedmetadata`,n)},[Be,tt]),(0,_.useEffect)(()=>{if(!E||!g||!v||Ne==null)return;if(Fe.length){it(Fe);return}let e=!1;return u.get(ue(v,Ne)).then(({data:t})=>{e||it(de(t))}).catch(()=>{}),()=>{e=!0}},[E,g,v,Ne,Fe]);let vt=e=>{if(!v||Ne==null)return;let t=Re.current;t&&(T.current=t.currentTime),et(e),nt(me(v,Ne,e)),Qe(!1),Ie?.(e)},yt=oe,bt=at?.download_speed??0,xt=at?.active_peers,St=at?.total_peers??0,Ct=at?.connected_seeders??0,wt=xt==null?`—`:`${xt}/${St}`,A=xt==null?`—`:`↑${Ct}`,Tt=ie(ot.Filled,ot.Capacity)??`—`,Et=v==null?null:(0,$.jsxs)(`div`,{className:`mt-0.5 flex min-w-0 flex-wrap items-center gap-x-2.5 gap-y-1 text-[11px] leading-none text-white/80 drop-shadow sm:text-xs`,"aria-label":`${S(`DownloadSpeed`)}, ${S(`Peers`)}, ${S(`CacheFilled`)}`,children:[(0,$.jsxs)(Wu,{tip:S(`DownloadSpeed`),children:[(0,$.jsx)(pe,{className:`size-3 shrink-0`,strokeWidth:2.25,"aria-hidden":!0}),(0,$.jsx)(`span`,{className:`tabular-nums`,children:re(bt)})]}),(0,$.jsxs)(Wu,{tip:S(`Peers`),children:[(0,$.jsx)(ae,{className:`size-3 shrink-0`,strokeWidth:2,"aria-hidden":!0}),(0,$.jsx)(`span`,{className:`tabular-nums`,children:wt}),(0,$.jsx)(`span`,{className:`tabular-nums text-white/70`,children:A})]}),(0,$.jsxs)(Wu,{tip:S(`CacheFilled`),children:[(0,$.jsx)(ee,{className:`size-3 shrink-0`,strokeWidth:2,"aria-hidden":!0}),(0,$.jsx)(`span`,{className:`min-w-0 truncate tabular-nums`,children:Tt})]})]}),j=O?(0,$.jsxs)(t,{isOpen:Ze,onOpenChange:Qe,children:[(0,$.jsx)(t.Trigger,{children:(0,$.jsx)(r,{isIconOnly:!0,variant:`ghost`,className:`${ge} text-white hover-fine:bg-white/15`,"aria-label":S(`SelectAudioTrack`),children:(0,$.jsx)(fe,{...yt,"aria-hidden":!0})})}),(0,$.jsxs)(t.Content,{placement:`bottom end`,className:`max-h-72 w-72 overflow-y-auto border border-white/10 bg-neutral-950/95 p-1 backdrop-blur-xl`,children:[(0,$.jsx)(`p`,{className:`px-2 py-1.5 text-[11px] font-semibold tracking-wide text-muted uppercase`,children:S(`SelectAudioTrack`)}),rt.map((e,t)=>{let{title:n,meta:i}=le(e,t);return(0,$.jsx)(r,{variant:t===$e?`secondary`:`ghost`,className:`h-auto w-full justify-start gap-2 py-2`,onPress:()=>vt(t),children:(0,$.jsxs)(`span`,{className:`min-w-0 flex-1 text-left`,children:[(0,$.jsx)(`span`,{className:`block truncate text-sm font-medium`,children:n}),i?(0,$.jsx)(`span`,{className:`mt-0.5 block truncate text-xs text-muted`,children:i}):null]})},t)})]})]}):null,M=C?null:(0,$.jsx)(r,{isIconOnly:!0,variant:`ghost`,className:`${ge} text-white hover-fine:bg-white/15`,"aria-label":S(Je?`ExitFullscreen`:`ExpandPlayer`),onPress:()=>Ye(e=>!e),children:Je?(0,$.jsx)(Ee,{...yt,"aria-hidden":!0}):(0,$.jsx)(te,{...yt,"aria-hidden":!0})});return(0,$.jsxs)($.Fragment,{children:[ye&&(De?(0,$.jsxs)(r,{variant:ke?`primary`:`secondary`,size:`sm`,onPress:ht,className:ke?`min-h-10 shrink-0 px-3`:`min-h-10 min-w-[72px] max-w-full flex-1`,children:[ke?(0,$.jsx)(he,{...se,fill:`currentColor`,"aria-hidden":!0}):null,S(`Play`)]}):(0,$.jsxs)(r,{variant:`secondary`,onPress:ht,children:[(0,$.jsx)(he,{...se,"aria-hidden":!0}),S(`Play`)]})),(0,$.jsx)(s,{state:_t,children:(0,$.jsx)(s.Backdrop,{isDismissable:!Ke,className:`bg-black/75 backdrop-blur-sm`,children:(0,$.jsx)(s.Container,{size:C?`full`:`lg`,scroll:`inside`,className:C?`ts-player-modal-full h-dvh p-0`:`p-4 sm:p-5`,children:(0,$.jsx)(s.Dialog,{className:C?`h-dvh max-h-dvh overflow-hidden rounded-none border-0 bg-black text-white shadow-none`:`overflow-hidden border border-white/10 bg-[#0a0e0c] text-white shadow-2xl shadow-black/50`,style:lt,children:(0,$.jsxs)(s.Body,{className:C?`flex h-full min-h-0 flex-col gap-0 p-0`:`gap-0 p-0`,children:[We?(0,$.jsxs)(`div`,{className:`space-y-3 p-4`,children:[(0,$.jsx)(o,{status:`danger`,children:S(`PlaybackError`)}),(0,$.jsx)(r,{variant:`secondary`,onPress:()=>window.open(n,`_blank`,`noopener,noreferrer`),children:S(`OpenLink`)})]}):(0,$.jsx)(Lu,{isMobile:C,showPip:k,showCaptionsButton:ct,className:C?`ts-player--mobile`:`ts-player--desktop`,style:C?{width:`100%`,height:`100%`}:{width:`100%`,maxHeight:ut,aspectRatio:`16 / 9`},topChrome:(0,$.jsxs)(`div`,{className:`flex items-start gap-2 px-3 pb-10 pt-[max(0.75rem,env(safe-area-inset-top))] pl-[max(0.75rem,env(safe-area-inset-left))] pr-[max(0.75rem,env(safe-area-inset-right))]`,children:[(0,$.jsxs)(`div`,{className:`min-w-0 flex-1`,children:[(0,$.jsx)(`p`,{className:`truncate text-sm font-semibold text-white drop-shadow`,title:a||S(`Play`),children:a||S(`Play`)}),Et]}),j,M,(0,$.jsx)(r,{isIconOnly:!0,variant:`ghost`,className:`${ge} text-white hover-fine:bg-white/15`,"aria-label":S(`Close`),onPress:gt,children:(0,$.jsx)(ce,{...yt,"aria-hidden":!0})})]}),extraControls:null,children:(0,$.jsx)(`video`,{slot:`media`,ref:dt,autoPlay:!0,playsInline:!0,crossOrigin:`anonymous`,className:`ts-player-video`,onTimeUpdate:()=>{pt(),He&&Ue(!1)},onLoadedMetadata:()=>{Ue(!1),Ge(!1),mt()},onWaiting:()=>Ue(!0),onCanPlay:()=>{Ue(!1),Ge(!1)},onPlaying:()=>Ue(!1),onPause:ft,onError:()=>{Ue(!1),Ge(!0)},children:!g&&Me?(0,$.jsx)(`track`,{kind:`captions`,src:Me,srcLang:`und`,label:`Captions`,default:!0}):null},tt)}),He&&!We?(0,$.jsx)(`div`,{className:`pointer-events-none absolute inset-0 z-10 grid place-items-center bg-black/55`,children:(0,$.jsxs)(`div`,{className:`flex flex-col items-center gap-3 px-4 text-center`,children:[(0,$.jsx)(l,{size:`lg`,color:`current`,className:`text-accent`}),(0,$.jsx)(`p`,{className:`text-sm font-medium text-white/90`,children:S(`Buffering`)})]})}):null]})})})})})]})}export{Ju as default};