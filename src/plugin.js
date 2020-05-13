/* global kiwi:true */

import md5 from 'md5';
import * as config from './config.js';

kiwi.plugin('gravatar', (kiwi) => {
    config.setDefaults();

    kiwi.on('irc.join', (event, net) => {
        kiwi.Vue.nextTick(() => {
            updateAvatar(net, event.nick);
        });
    });

    kiwi.on('irc.wholist', (event, net) => {
        let nicks = event.users.map((user) => user.nick);
        kiwi.Vue.nextTick(() => {
            nicks.forEach((nick) => {
                updateAvatar(net, nick, false);
            });
        });
    });

    kiwi.on('irc.account', (event, net) => {
        kiwi.Vue.nextTick(() => {
            updateAvatar(net, event.nick, true);
        });
    });

    function updateAvatar(net, nick, _force) {
        let force = !!_force;
        let user = kiwi.state.getUser(net.id, nick);
        if (!user) {
            return;
        }

        if (!force && user.avatar && user.avatar.small) {
            return;
        }

        let url = (user.account) ?
            kiwi.state.setting('plugin-gravatar.gatewayURL') + 'gravatar/' + user.account :
            kiwi.state.setting('plugin-gravatar.gravatarURL') + md5(net.name + ':' + user.nick);

        console.log('setting avatar for', user.nick);
        setAvatar(user, url);
    }

    function setAvatar(user, _url) {
        let url = _url;

        let gRating = kiwi.state.setting('plugin-gravatar.gravatarRating');
        let gFallback = kiwi.state.setting('plugin-gravatar.gravatarFallback');

        url += '?r=' + gRating;
        url += '&d=' + gFallback;

        user.avatar.small = url + '&s=30';
        user.avatar.large = url + '&s=200';
    }
});
