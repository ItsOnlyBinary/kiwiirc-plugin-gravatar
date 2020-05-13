/* global kiwi:true */

export function setDefaults() {
    setSettingDefault('plugin-gravatar.gatewayURL', '/');
    setSettingDefault('plugin-gravatar.gravatarURL', '//www.gravatar.com/avatar/');
    setSettingDefault('plugin-gravatar.gravatarRating', 'g');
    setSettingDefault('plugin-gravatar.gravatarFallback', 'robohash');
}

function setSettingDefault(name, value) {
    if (kiwi.state.getSetting('settings.' + name) === undefined) {
        kiwi.state.setSetting('settings.' + name, value);
    }
}
