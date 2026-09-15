ts.web.staticFile("/", "index.html");
ts.web.staticFile("/index.html", "index.html");
ts.web.staticDir("/css", "css");
ts.web.staticDir("/js", "js");
ts.web.staticDir("/img", "img");
ts.web.staticDir("/i18n", "i18n");

console.log("[Default UI] Static routes registered");