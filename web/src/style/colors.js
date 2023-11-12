import { rgba } from 'polished'

export const themeColors = {
  light: {
    app: {
      headerToggleColor: '#4db380',
      appSecondaryColor: '#cbe8d9',
      sidebarBGColor: '#575757',
      sidebarFillColor: '#dee3e5',
      paperColor: '#eeeeee',
    },
    torrentCard: {
      accentCardColor: '#337a57',
      buttonBGColor: rgba('#337a57', 0.5),
      cardPrimaryColor: '#00a572',
      cardSecondaryColor: '#74c39c',
    },
    dialogTorrentDetailsContent: {
      posterBGColor: '#74c39c',
      gradientStartColor: '#e4f6ed',
      gradientEndColor: '#b5dec9',
      chacheSectionBGColor: '#88cdaa',
      widgetFontColor: '#fff',
      titleFontColor: '#000',
      subNameFontColor: '#7c7b7c',
      torrentFilesSectionBGColor: '#f1eff3',
    },
    detailedView: {
      gradientStartColor: '#e4f6ed',
      gradientEndColor: '#b5dec9',
      cacheSectionBGColor: '#fff',
    },
    addDialog: {
      gradientStartColor: '#e4f6ed',
      gradientEndColor: '#b5dec9',
      fontColor: '#000',
      notificationErrorBGColor: '#cda184',
      notificationSuccessBGColor: '#88cdaa',
      languageSwitchBGColor: '#74c39c',
      languageSwitchFontColor: '#e4f6ed',
      posterBGColor: '#74c39c',
    },
    torrentFunctions: {
      fontColor: '#000',
    },
    table: {
      defaultPrimaryColor: '#009879',
      defaultSecondaryColor: '#00a383',
      defaultTertiaryColor: '#03aa89',
    },
    settingsDialog: {
      contentBG: '#f1eff3',
      footerBG: '#fff',
    },
  },
  dark: {
    app: {
      headerToggleColor: '#545a5e',
      appSecondaryColor: '#545a5e',
      sidebarBGColor: '#323637',
      sidebarFillColor: '#dee3e5',
      paperColor: '#323637',
    },
    torrentCard: {
      accentCardColor: '#323637',
      buttonBGColor: rgba('#323637', 0.5),
      cardPrimaryColor: '#545a5e',
      cardSecondaryColor: rgba('#dee3e5', 0.4),
    },
    dialogTorrentDetailsContent: {
      posterBGColor: rgba('#dee3e5', 0.4),
      gradientStartColor: '#656f76',
      gradientEndColor: '#545a5e',
      chacheSectionBGColor: '#3c4244',
      widgetFontColor: rgba('#fff', 0.8),
      titleFontColor: '#f1eff3',
      subNameFontColor: '#dee3e5',
      torrentFilesSectionBGColor: rgba('#545a5e', 0.9),
    },
    detailedView: {
      gradientStartColor: '#656f76',
      gradientEndColor: '#545a5e',
      cacheSectionBGColor: '#949ca0',
    },
    addDialog: {
      gradientStartColor: '#656f76',
      gradientEndColor: '#545a5e',
      fontColor: '#fff',
      notificationErrorBGColor: '#c82e3f',
      notificationSuccessBGColor: '#323637',
      languageSwitchBGColor: '#545a5e',
      languageSwitchFontColor: '#dee3e5',
      posterBGColor: '#dee3e5',
    },
    torrentFunctions: {
      fontColor: '#f1eff3',
    },
    table: {
      defaultPrimaryColor: '#323637',
      defaultSecondaryColor: rgba('#545a5e', 0.9),
      defaultTertiaryColor: '#545a5e',
    },
    settingsDialog: {
      contentBG: '#5a6166',
      footerBG: '#323637',
    },
  },
}

export const mainColors = {
  light: {
    primary: '#00a572',
    secondary: '#00a572',
    labels: rgba('#000', 0.9),
  },
  dark: {
    primary: '#323637',
    secondary: '#dee3e5',
    labels: rgba('#fff', 0.9),
  },
}
