/******************************************************************************/
//Tizen Interaction Plugin v0.0.1
//(c) 2021 Benjamin Zachey
//Action syntax example:
//- "content:request:interaction:init@http://msx.benzac.de/interaction/tizen.html"
/******************************************************************************/

/******************************************************************************/
//TizenPlayer
/******************************************************************************/
function TizenPlayer() {
  var pushToArray = function(array, items) {
    if (array != null && items != null) {
      if (Array.isArray(items)) {
        for (var i = 0; i < items.length; i++) {
          array.push(items[i]);
        }
      } else {
        array.push(items);
      }
    }
  };
  var getValueLabel = function(value) {
    if (typeof value == "string") {
      return TVXTools.strFullCheck(value, "-");
    }
    return value != null ? TVXTools.strValue(value) : "-";
  };
  var getPropertyLabel = function(value, unit) {
    if (TVXTools.isFullStr(value) && value.indexOf("|") >= 0) {
      return value.split("|")[0];
    }
    if (value != null) {
      if (unit != null) {
        return value + " " + unit;
      }
      return TVXTools.strValue(value);
    }
    return "Unknown";
  };
  var getPropertyValue = function(value) {
    if (TVXTools.isFullStr(value) && value.indexOf("|") >= 0) {
      value = value.split("|")[1];
    }
    if (TVXTools.isFullStr(value) && value.indexOf("num:") == 0) {
      return TVXTools.strToNum(value.substr(4));
    }
    return value;
  };
  var getTrackLabel = function(track) {
    if (track != null) {
      var prefix = "Track " + track.index;
      var suffix = track.info != null ? track.info.language : null;//Audio track
      if (suffix == null) {
        suffix = track.info != null ? track.info.track_lang : null;//Text track
      }
      return TVXTools.isFullStr(suffix) ? prefix + " (" + suffix + ")" : prefix;
    }
    return "None";
  };
  var createPropertyControls = function(y, propertyIcon, propertyLabel, propertyKey, propertyValue, propertyUnit, availableValues, nextButton) {
    var panelItems = [];
    var selectedPropertyLabel = getPropertyLabel(propertyValue, propertyUnit);
    var firstPropertyValue = null;
    var nextPropertyValue = null;
    var selectNext = false;
    if (availableValues != null) {
      for (var i = 0; i < availableValues.length; i++) {
        var value = getPropertyValue(availableValues[i]);
        var label = getPropertyLabel(availableValues[i], propertyUnit);
        var selected = value === propertyValue;
        if (firstPropertyValue == null) {
          firstPropertyValue = value;
        }
        if (selected) {
          selectNext = true;
          selectedPropertyLabel = label;
        } else if (selectNext) {
          selectNext = false;
          nextPropertyValue = value;
        }
        panelItems.push({
          focus: selected,
          extensionIcon: selected ? "check" : "blank",
          label: label,
          action: selected ? "back" : "[invalidate:content|back|player:commit]",
          data: {
            key: propertyKey,
            value: value,
            action: "reload:content"
          }
        });
      }
    }
    if (nextPropertyValue == null) {
      nextPropertyValue = firstPropertyValue;
    }
    return [{
        enable: panelItems.length > 0,
        type: "control",
        layout: "0," + y + "," + (nextButton ? "7,1" : "8,1"),
        icon: propertyIcon,
        label: propertyLabel,
        extensionLabel: selectedPropertyLabel,
        action: "panel:data",
        data: {
          headline: propertyLabel,
          compress: panelItems.length > 6,
          template: {
            type: "control",
            enumerate: false,
            layout: panelItems.length > 8 ? "0,0,5,1" : panelItems.length > 6 ? "0,0,10,1" : "0,0,8,1"
          },
          items: panelItems
        }
      }, {
        display: nextButton,
        enable: nextPropertyValue != null,
        type: "button",
        icon: "navigate-next",
        iconSize: "small",
        layout: "7," + y + ",1,1",
        action: "[invalidate:content|player:commit]",
        data: {
          key: propertyKey,
          value: nextPropertyValue,
          action: "reload:content"
        }
      }];
  };
  var createTrackControl = function(y, propertyIcon, propertyLabel, propertyKey, currentTrack, availableTracks) {
    var panelItems = [];
    if (availableTracks != null) {
      for (var i = 0; i < availableTracks.length; i++) {
        var track = availableTracks[i];
        var selected = track.index === (currentTrack != null ? currentTrack.index : -1);
        panelItems.push({
          focus: selected,
          extensionIcon: selected ? "check" : "blank",
          label: getTrackLabel(track),
          action: selected ? "back" : "[invalidate:content|back|player:commit]",
          data: {
            key: propertyKey,
            value: track.index,
            action: "reload:content"
          }
        });
      }
    }
    return {
      enable: panelItems.length > 0,
      type: "control",
      layout: "0," + y + ",8,1",
      icon: propertyIcon,
      label: propertyLabel,
      extensionLabel: getTrackLabel(currentTrack),
      action: "panel:data",
      data: {
        headline: propertyLabel,
        compress: panelItems.length > 6,
        template: {
          type: "control",
          enumerate: false,
          layout: panelItems.length > 8 ? "0,0,5,1" : panelItems.length > 6 ? "0,0,10,1" : "0,0,8,1"
        },
        items: panelItems
      }
    };
  };
  var createControlItems = function(infoData) {
    var items = [];
    pushToArray(items, createPropertyControls(0, "featured-video", "Display Area", "tizen:display:area",
        infoData && infoData.display != null ? infoData.display.area : null, null, [
          "21x9|0,0.119,1,0.762",
          "16x9 (Default)|0,0,1,1",
          "4x3|0.125,0,0.75,1"
        ], true));
    pushToArray(items, createPropertyControls(1, "aspect-ratio", "Display Mode", "tizen:display:mode",
        infoData != null && infoData.display != null ? infoData.display.mode : null, null, [
          "Fit Screen (Default)|PLAYER_DISPLAY_MODE_LETTER_BOX",
          "Fill Screen|PLAYER_DISPLAY_MODE_FULL_SCREEN",
          "Auto Aspect Ratio|PLAYER_DISPLAY_MODE_AUTO_ASPECT_RATIO"
        ], true));
    pushToArray(items, createPropertyControls(2, "av-timer", "Buffer Timeout", "tizen:buffer:timeout",
        infoData != null && infoData.buffer != null ? infoData.buffer.timeout : null, "sec", [
          0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 15, "20 sec (Default)|num:20", 25, 30, 40
        ], false));
    pushToArray(items, createPropertyControls(3, "timelapse", "Buffer Size (Init)", "tizen:buffer:size:init",
        infoData != null && infoData.buffer != null && infoData.buffer.size != null ? infoData.buffer.size.init : null, "sec", null, false));
    pushToArray(items, createPropertyControls(4, "timelapse", "Buffer Size (Resume)", "tizen:buffer:size:resume",
        infoData != null && infoData.buffer != null && infoData.buffer.size != null ? infoData.buffer.size.resume : null, "sec", null, false));
    pushToArray(items, createTrackControl(5, "audiotrack", "Audio Track", "tizen:track:audio",
        infoData != null && infoData.stream != null ? infoData.stream.audio : null,
        infoData != null && infoData.tracks != null ? infoData.tracks.audio : null));
    pushToArray(items, createTrackControl(6, "subtitles", "Text Track", "tizen:track:text",
        infoData != null && infoData.stream != null ? infoData.stream.text : null,
        infoData != null && infoData.tracks != null ? infoData.tracks.text : null));
    return items;
  };
  var createInfoItems = function(infoData) {
    var infoKeys = "-";
    var infoValues = "-";
    if (infoData != null) {
      infoKeys = "Version:{br}State:{br}{br}";
      infoValues = getValueLabel(infoData.version) + "{br}" +
          getValueLabel(infoData.state) + "{br}{br}";
      if (infoData.stream != null) {
        if (infoData.stream.video != null && infoData.stream.video.info != null) {
          for (var key in infoData.stream.video.info) {
            infoKeys += "Video [" + key + "]:{br}";
            infoValues += getValueLabel(infoData.stream.video.info[key]) + "{br}";
          }
          infoKeys += "{br}";
          infoValues += "{br}";
        }
        if (infoData.stream.audio != null && infoData.stream.audio.info != null) {
          for (var key in infoData.stream.audio.info) {
            infoKeys += "Audio [" + key + "]:{br}";
            infoValues += getValueLabel(infoData.stream.audio.info[key]) + "{br}";
          }
          infoKeys += "{br}";
          infoValues += "{br}";
        }
        if (infoData.stream.text != null && infoData.stream.text.info != null) {
          for (var key in infoData.stream.text.info) {
            infoKeys += "Text [" + key + "]:{br}";
            infoValues += getValueLabel(infoData.stream.text.info[key]) + "{br}";
          }
        }
      }
    }
    return [{
        type: "space",
        layout: "8,0,3,8",
        offset: "0.25,0,0,0",
        truncation: "text",
        text: infoKeys
      }, {
        type: "space",
        layout: "11,0,5,8",
        offset: "0.25,0,-0.25,0",
        truncation: "text",
        text: "{col:msx-white}" + infoValues,
        live: {
          type: "airtime",
          duration: 2000,
          over: {
            action: "reload:content"
          }
        }
      }];
  };
  var createContentItems = function(infoData) {
    var items = [];
    pushToArray(items, createControlItems(infoData));
    pushToArray(items, createInfoItems(infoData));
    return items;
  };
  var createContent = function(infoData) {
    return {
      cache: false,
      compress: true,
      type: "pages",
      headline: "Tizen Player",
      extension: "{ico:msx-white:timer} " + TVXDateTools.getTimestamp(),
      pages: [{
          items: createContentItems(infoData)
        }]
    };
  };
  var createWarningContent = function(playerInfo) {
    return {
      type: "pages",
      headline: "Tizen Player",
      pages: [{
          items: [{
              type: "default",
              layout: "0,0,12,6",
              color: "msx-glass",
              headline: "{ico:msx-yellow:warning} Player Not Available",
              text: "Tizen player is required for this plugin. Current player is: {txt:msx-white:" + playerInfo + "}."
            }]
        }]
    };
  };
  var createDummyData = function() {
    return {
      version: "1.0",
      state: "PLAYING",
      display: {
        area: "0,0,1,1",
        mode: "PLAYER_DISPLAY_MODE_LETTER_BOX"
      },
      buffer: {
        timeout: 20,
        size: {
          init: 10,
          resume: 10
        }
      },
      stream: {
        video: {
          index: 0,
          info: {
            fourCC: "h264",
            Width: 1280,
            Height: 720,
            Bit_rate: 128000
          }
        }
      }
    };
  };
  this.handleRequest = function(playerInfo, dataId, callback) {
    if (dataId == "init") {
      if (TVXTools.isFullStr(playerInfo) && (playerInfo == "tizen" || playerInfo.indexOf("tizen/") == 0)) {
        TVXInteractionPlugin.requestPlayerResponse("tizen:info", function(data) {
          callback(createContent(data.response != null && data.response.tizen != null ? data.response.tizen.info : null));
        });
      } else {
        callback(createWarningContent(playerInfo));
      }
      return true;
    } else if (dataId == "dummy") {
      callback(createContent(createDummyData()));
      return true;
    }
    return false;
  };
}
/******************************************************************************/

/******************************************************************************/
//TizenHandler
/******************************************************************************/
function TizenHandler() {
  var playerInfo = null;
  var readyService = new TVXBusyService();
  var player = new TizenPlayer();

  this.ready = function() {
    readyService.start();
    TVXInteractionPlugin.requestData("info:base", function(data) {
      playerInfo = TVXTools.strFullCheck(data.info != null ? data.info.player : null, "unknown");
      readyService.stop();
    });
  };
  this.handleRequest = function(dataId, data, callback) {
    readyService.onReady(function() {
      if (!player.handleRequest(playerInfo, dataId, callback)) {
        callback();
      }
    });
  };
}
/******************************************************************************/

/******************************************************************************/
//Setup
/******************************************************************************/
window.onload = function() {
  TVXInteractionPlugin.setupHandler(new TizenHandler());
  TVXInteractionPlugin.init();
};
/******************************************************************************/