package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Структура для сайта из data.json
type SiteInfo struct {
	Name      string      `json:"name"`
	BaseURL   string      `json:"base_url"`
	URLProbe  string      `json:"url_probe"` // Используем, если есть, для проверки
	ErrorType string      `json:"errorType"`
	ErrorCode interface{} `json:"errorCode"` // Может быть int или string
	ErrorMsg  string      `json:"errorMsg"`
	// Добавим поле для User-Agent, если понадобится
	// UserAgent string `json:"user_agent,omitempty"`
}

// Вместо embed просто определяем минимальный набор сайтов для проверки
// Это позволит избежать проблем сборки
const sitesDataStr = `[
{
        "name": "Instagram",
        "base_url": "https://instagram.com/{}",
        "url_probe": "https://imginn.com/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg":"<title>Page Not Found - imginn.com</title>"
      },
      {
        "name": "Twitter/X",
        "base_url": "https://twitter.com/{}",
        "follow_redirects": true,
        "errorType": "unknown"
      },
      {
        "name": "GitHub",
        "base_url": "https://github.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Reddit",
        "base_url": "https://www.reddit.com/user/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<title>Reddit - Dive into anything</title>"
      },
      {
        "name": "Facebook",
        "base_url": "https://www.facebook.com/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<title>Facebook</title>"
      },
      {
        "name": "YouTube",
        "base_url": "https://www.youtube.com/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "TikTok",
        "base_url": "https://www.tiktok.com/@{}",
        "follow_redirects": true,
        "errorType": "profilePresence",
        "errorMsg": "shareMeta"
      },
      {
        "name": "About Me",
        "base_url": "https://about.me/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Independent Academia",
        "base_url": "https://independent.academia.edu/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Airbit",
        "base_url": "https://airbit.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Airliners",
        "base_url": "https://www.airliners.net/user/{}/profile/photos",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Duolingo",
        "base_url": "https://www.duolingo.com/profile/{}",
        "url_probe": "https://www.duolingo.com/2017-06-30/users?username={}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "{\"users\":[]}"
      },
      {
        "name": "Pinterest",
        "base_url": "https://www.pinterest.com/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<title></title>"
      },
      {
        "name": "VSCO",
        "base_url": "https://vsco.co/{}/gallery",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Snapchat",
        "base_url": "https://www.snapchat.com/add/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Threads",
        "base_url": "https://www.threads.net/{}",
        "follow_redirects": true,
        "errorType": "unknown",
        "errorCode": 200
      },
      {
        "name": "Tumblr",
        "base_url": "https://www.tumblr.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Keybase",
        "base_url": "https://keybase.io/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Wattpad",
        "base_url": "https://www.wattpad.com/user/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Mastodon Social",
        "base_url": "https://mastodon.social/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "MSTDN Social",
        "base_url": "https://mstdn.social/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Mas.to",
        "base_url": "https://mas.to/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Mastodon World",
        "base_url": "https://mastodon.world/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Fosstodon",
        "base_url": "https://fosstodon.org/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Hachyderm",
        "base_url": "https://hachyderm.io/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Vivaldi Social",
        "base_url": "https://social.vivaldi.net/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Techhub Social",
        "base_url": "https://techhub.social/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Wakatime",
        "base_url": "https://wakatime.com/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Twitch",
        "base_url": "https://twitch.tv/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<meta property='og:description' content='Twitch is the world&#39;s leading video platform and community for gamers.'>"
      },
      {
        "name": "Rumble",
        "base_url": "https://rumble.com/c/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Kick",
        "base_url": "https://kick.com/{}",
        "follow_redirects": true,
        "errorType": "unknown"
      },
      {
        "name": "Yandex Dzen",
        "base_url": "https://dzen.ru/{}",
        "follow_redirects": true,
        "errorType": "status_code",
        "cookies": [
          {
            "name": "zen_sso_checked",
            "value": "1"
          }
        ]
      },
      {
        "name": "YVision KZ",
        "base_url": "https://yvision.kz/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Giphy",
        "base_url": "https://giphy.com/channel/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Bluesky",
        "base_url": "https://bsky.app/profile/{}.bsky.social",
        "url_probe": "https://public.api.bsky.app/xrpc/app.bsky.actor.getProfile?actor={}.bsky.social",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "9GAG",
        "base_url": "https://9gag.com/u/{}",
        "follow_redirects": true,
        "errorType": "status_code",
        "cookies": [
          {
            "name": "ts1",
            "value": "4b134fab9439a52a1a8d1789265b0404523ccb08"
          }
        ]
      },
      {
        "name": "Flickr",
        "base_url": "https://flickr.com/photos/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Behance",
        "base_url": "https://behance.net/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Buy Me a Coffee",
        "base_url": "https://buymeacoffee.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Ko-fi",
        "base_url": "https://ko-fi.com/{}",
        "follow_redirects": false,
        "errorType": "status_code",
        "errorCode": 302
      },
      {
        "name": "Tinder",
        "base_url": "https://tinder.com/@{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<title data-react-helmet=\"true\">Tinder | Dating, Make Friends &amp; Meet New People</title>"
      },
      {
        "name": "LinkedIn",
        "base_url": "https://www.linkedin.com/in/{}",
        "follow_redirects": true,
        "errorType": "unknown"
      },
      {
        "name": "Vimeo",
        "base_url": "https://vimeo.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Patreon",
        "base_url": "https://www.patreon.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Substack",
        "base_url": "https://{}.substack.com",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Medium",
        "base_url": "https://medium.com/@{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<title data-rh=\"true\">Medium</title>"
      },
      {
        "name": "DEV Community",
        "base_url": "https://dev.to/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Hashnode",
        "base_url": "https://hashnode.com/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Spotify",
        "base_url": "https://open.spotify.com/user/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Clubhouse",
        "base_url": "https://www.joinclubhouse.com/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Foursquare",
        "base_url": "https://foursquare.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "SoundCloud",
        "base_url": "https://soundcloud.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Vero",
        "base_url": "https://vero.co/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "_not-found-page-container_onczy_1"
      },
      {
        "name": "Figma",
        "base_url": "https://www.figma.com/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Linktree",
        "base_url": "https://www.linktr.ee/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "\"statusCode\":404"
      },
      {
        "name": "Beacons.ai",
        "base_url": "https://beacons.ai/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "Beacons | Mobile Websites for Creators"
      },
      {
        "name": "Bio.link",
        "base_url": "https://bio.link/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Milkshake",
        "base_url": "https://msha.ke/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Snipfeed",
        "base_url": "https://snipfeed.co/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Ayo.so",
        "base_url": "https://ayo.so/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Carrd",
        "base_url": "https://{}.carrd.co",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Steam Community (User)",
        "base_url": "https://steamcommunity.com/id/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<title>Steam Community :: Error</title>"
      },
      {
        "name": "Dev Community",
        "base_url": "https://dev.to/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Daily.dev",
        "base_url": "https://app.daily.dev/{}",
        "follow_redirects": true,
        "errorType": "profilePresence",
        "errorMsg": "{\"props\":{\"pageProps\":{\"user\":{\"id\":"
      },
      {
        "name": "HackerNews",
        "base_url": "https://news.ycombinator.com/user?id={}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "No such user."
      },
      {
        "name": "HackTheBox Forum",
        "base_url": "https://forum.hackthebox.com/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "1337x.to",
        "base_url": "https://www.1337x.to/user/{}/",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<title>Error something went wrong.</title>"
      },
      {
        "name": "7Cups",
        "base_url": "https://www.7cups.com/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "8Tracks",
        "base_url": "https://8tracks.com/{}",
        "url_probe": "https://8tracks.com/users/check_username?login={}&format=jsonh",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "{\"available\":true"
      },
      {
        "name": "All My Links",
        "base_url": "https://allmylinks.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Aniworld.to",
        "base_url": "https://aniworld.to/user/profil/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<title>Profil | AniWorld.to - Animes gratis legal online ansehen</title>"
      },
      {
        "name": "Anilist",
        "base_url": "https://anilist.co/user/{}/",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<title>AniList</title>"
      },
      {
        "name": "Apple Developers",
        "base_url": "https://developer.apple.com/forums/profile/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Apple Discussions",
        "base_url": "https://discussions.apple.com/profile/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Archive Of Our Own (AO3)",
        "base_url": "https://archiveofourown.org/users/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Telegram",
        "base_url": "https://t.me/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "If you have <strong>Telegram</strong>, you can contact"
      },
      {
        "name": "LastFM",
        "base_url": "https://last.fm/user/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Chess",
        "base_url": "https://www.chess.com/member/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Lichess",
        "base_url": "https://lichess.org/@/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Codecademy",
        "base_url": "https://www.codecademy.com/profiles/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<title>Profile Not Found | Codecademy</title>"
      },
      {
        "name": "Gitlab",
        "base_url": "https://gitlab.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
          "name": "sourcehut",
          "base_url": "https://sr.ht/~{}/",
          "follow_redirects": true,
          "errorType": "status_code"
      },
      {
        "name": "Disqus",
        "base_url": "https://disqus.com/by/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Docker Hub",
        "base_url": "https://hub.docker.com/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Kali Linux Forums",
        "base_url": "https://forums.kali.org/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Imgur",
        "base_url": "https://imgur.com/user/{}",
        "url_probe":"https://api.imgur.com/account/v1/accounts/{}?client_id=546c25a59c58ad7",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "GameFAQs Community",
        "base_url": "https://gamefaqs.gamespot.com/community/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "KASKUS",
        "base_url": "https://www.kaskus.co.id/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "TripAdvisor Forums",
        "base_url": "https://www.tripadvisor.com/Profile/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Lobsters",
        "base_url": "https://lobste.rs/~{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "PHUCKS",
        "base_url": "https://phuks.co/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Hubski",
        "base_url": "https://hubski.com/user/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<a href=\"/\">No such user.</a>"
      },
      {
        "name": "Tildes",
        "base_url": "https://tildes.net/~{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Leetcode",
        "base_url": "https://leetcode.com/u/{}",
        "follow_redirects": true,
        "errorType": "unknown"
      },
      {
        "name": "SourceForge",
        "base_url": "https://sourceforge.net/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Bitwarden Forums",
        "base_url": "https://community.bitwarden.com/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Blipfoto",
        "base_url": "https://www.blipfoto.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Archive.org",
        "base_url": "https://archive.org/details/@{}",
        "follow_redirects": true,
        "errorType": "unknown",
        "errorCode": 200
      },
      {
        "name": "ArtStation",
        "base_url": "https://www.artstation.com/{}",
        "follow_redirects": true,
        "errorType": "unknown"
      },
      {
        "name": "Asciinema",
        "base_url": "https://asciinema.org/~{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Fedora Discussion",
        "base_url": "https://discussion.fedoraproject.org/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Atcoder",
        "base_url": "https://atcoder.jp/users/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Audio Jungle",
        "base_url": "https://audiojungle.net/user/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Autofrage",
        "base_url": "https://www.autofrage.net/nutzer/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Avizo",
        "base_url": "https://www.avizo.cz/{}",
        "follow_redirects": false,
        "errorType": "status_code",
        "errorCode": 301
      },
      {
        "name": "BOOTH",
        "base_url": "https://{}.booth.pm/",
        "follow_redirects": false,
        "errorType": "status_code",
        "errorCode": 302
      },
      {
        "name": "Bandcamp",
        "base_url": "https://www.bandcamp.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Blogger",
        "base_url": "https://{}.blogspot.com",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "BoardGameGeek",
        "base_url": "https://boardgamegeek.com/user/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "User not found"
      },
      {
        "name": "Bookcrossing",
        "base_url": "https://www.bookcrossing.com/mybookshelf/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Brave Community",
        "base_url": "https://community.brave.com/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Strava",
        "base_url": "https://www.strava.com/athletes/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Bugcrowd",
        "base_url": "https://bugcrowd.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Buzzfeed",
        "base_url": "https://www.buzzfeed.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "CGTrader",
        "base_url":"https://www.cgtrader.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "CNET",
        "base_url":"https://www.cnet.com/profiles/{}/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "CSSBattle",
        "base_url":"https://cssbattle.dev/player/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<title>CSSBattle</title>"
      },
      {
        "name": "CTAN",
        "base_url":"https://ctan.org/author/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Caddy Community",
        "base_url":"https://caddy.community/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
       "name": "Car Talk Community",
        "base_url":"https://community.cartalk.com/u/{}",
        "follow_redirects": true,
        "errorType": "status_code" 
      },
      {
        "name": "Career.habr",
        "base_url": "https://career.habr.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Championat",
        "base_url": "https://www.championat.com/user/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Chaos",
        "base_url": "https://chaos.social/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Chatujme.cz",
        "base_url": "https://profil.chatujme.cz/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Choice Community",
        "base_url": "https://choice.community/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Clapper",
        "base_url": "https://clapperapp.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Cloudflare Community",
        "base_url": "https://community.cloudflare.com/u/{}",
        "follow_redirects": true,
        "errorType": "unknown"
      },
      {
        "name": "Clubhouse",
        "base_url": "https://www.clubhouse.com/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Code Snippet Wiki",
        "base_url": "https://codesnippets.fandom.com/wiki/User:{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Codeberg",
        "base_url": "https://codeberg.org/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Codecademy",
        "base_url": "https://www.codecademy.com/profiles/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "This profile could not be found"
      },
      {
        "name": "Codechef",
        "base_url": "https://www.codechef.com/users/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<meta property=\"og:title\" content=\"CodeChef | CodeChef: Practical coding for everyone\" />"
      },
      {
        "name": "Codeforces",
        "base_url": "https://codeforces.com/profile/{}",
        "url_probe": "https://codeforces.com/api/user.info?handles={}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Codepen",
        "base_url": "https://codepen.io/{}",
        "follow_redirects": true,
        "errorType": "unknown"
      },
      {
        "name": "Coders Rank",
        "base_url": "https://profile.codersrank.io/user/{}/",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "not a registered member"
      },
      {
        "name": "Coderwall",
        "base_url": "https://coderwall.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Codewars",
        "base_url": "https://www.codewars.com/users/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "ColourLovers",
        "base_url": "https://www.colourlovers.com/lover/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Coroflot",
        "base_url": "https://www.coroflot.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Cracked",
        "base_url": "https://www.cracked.com/members/{}",
        "follow_redirects": false,
        "errorType": "status_code",
        "errorCode": 302
      },
      {
        "name": "Crevado",
        "base_url": "https://{}.crevado.com",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Crowdin",
        "base_url": "https://crowdin.com/profile/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Cryptomator Forum",
        "base_url": "https://community.cryptomator.org/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Cults3D",
        "base_url": "https://cults3d.com/en/users/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "CyberDefenders",
        "base_url": "https://cyberdefenders.org/p/{}",
        "follow_redirects": false,
        "errorType": "status_code",
        "errorCode": 301
      },
      {
        "name": "DMOJ",
        "base_url": "https://dmoj.ca/user/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "DailyMotion",
        "base_url": "https://www.dailymotion.com/{}",
        "follow_redirects": true,
        "errorType": "unknown",
        "errorCode": 200
      },
      {
        "name": "Dealabs",
        "base_url": "https://www.dealabs.com/profile/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "DeviantART",
        "base_url": "https://{}.deviantart.com",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Discogs",
        "base_url": "https://www.discogs.com/user/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Eintracht Frankfurt Forum",
        "base_url": "https://community.eintracht.de/fans/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Envato Forum",
        "base_url": "https://forums.envato.com/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Exposure",
        "base_url": "https://{}.exposure.co/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Exophase",
        "base_url": "https://www.exophase.com/user/{}/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "EyeEm",
        "base_url": "https://www.eyeem.com/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Fameswap",
        "base_url": "https://fameswap.com/user/{}",
        "follow_redirects": true,
        "errorType": "unknown"
      },
      {
        "name": "Fanpop",
        "base_url": "https://www.fanpop.com/fans/{}",
        "follow_redirects": false,
        "errorType": "status_code",
        "errorCode": 302
      },
      {
        "name": "Finanzfrage",
        "base_url": "https://www.finanzfrage.net/nutzer/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Flightradar24",
        "base_url": "https://my.flightradar24.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Flipboard",
        "base_url": "https://flipboard.com/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Rusfootball",
        "base_url": "https://www.rusfootball.info/user/{}/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "FortniteTracker",
        "base_url": "https://fortnitetracker.com/profile/all/{}",
        "follow_redirects": true,
        "errorType": "unknown"
      },
      {
        "name": "Freelance.habr",
        "base_url": "https://freelance.habr.com/freelancers/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Freelancer",
        "base_url": "https://www.freelancer.com/u/{}",
        "follow_redirects":true,
        "errorType": "status_code"
      },
      {
        "name": "Freesound",
        "base_url": "https://freesound.org/people/{}/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "GaiaOnline",
        "base_url": "https://www.gaiaonline.com/profiles/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "No user ID specified or user does not exist!"
      },
      {
        "name": "Gamespot",
        "base_url": "https://www.gamespot.com/profile/{}/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "GeeksforGeeks",
        "base_url": "https://auth.geeksforgeeks.org/user/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Genius (Artists)",
        "base_url": "https://genius.com/artists/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Genius (Users)",
        "base_url": "https://genius.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Gesundheitsfrage",
        "base_url": "https://www.gesundheitsfrage.net/nutzer/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "GetMyUni",
        "base_url": "https://www.getmyuni.com/author/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Giant Bomb",
        "base_url": "https://www.giantbomb.com/profile/{}/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "GitBook",
        "base_url": "https://{}.gitbook.io/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "GitHub Pages",
        "base_url": "https://{}.github.io/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Gitea",
        "base_url": "https://gitea.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Gitee",
        "base_url": "https://gitee.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "GoodReads",
        "base_url": "https://www.goodreads.com/{}",
        "follow_redirects":true,
        "errorType": "status_code"
      },
      {
        "name": "Gradle",
        "base_url": "https://plugins.gradle.org/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Grailed",
        "base_url": "https://www.grailed.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Gravatar",
        "base_url": "http://en.gravatar.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Gumroad",
        "base_url": "https://{}.gumroad.com/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Gutefrage",
        "base_url": "https://www.gutefrage.net/nutzer/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Hackaday",
        "base_url": "https://hackaday.io/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "HackenProof",
        "base_url": "https://hackenproof.com/hackers/{}",
        "follow_redirects": true,
        "errorType": "unknown"
      },
      {
        "name": "HackerEarth",
        "base_url": "https://hackerearth.com/@{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<title> 404 | HackerEarth</title>"
      },
      {
        "name": "HackerOne",
        "base_url": "https://hackerone.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "HackerRank",
        "base_url": "https://hackerrank.com/{}",
        "follow_redirects": true,
        "errorType": "unknown",
        "errorCode": 200
      },
      {
        "name": "Harvard Scholar",
        "base_url": "https://scholar.harvard.edu/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Hashnode",
        "base_url": "https://hashnode.com/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Houzz",
        "base_url": "https://houzz.com/user/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "HubPages",
        "base_url": "https://hubpages.com/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Hubski",
        "base_url": "https://hubski.com/user/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "No such user."
      },
      {
        "name": "IFTTT",
        "base_url": "https://www.ifttt.com/p/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "IRC-Galleria",
        "base_url": "https://irc-galleria.net/user/{}",
        "follow_redirects": false,
        "errorType": "status_code",
        "errorCode": 302
      },
      {
        "name": "Icons8 Community",
        "base_url": "https://community.icons8.com/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Guns.lol",
        "base_url": "https://guns.lol/{}",
        "follow_redirects": true,
        "errorType": "unknown",
        "errorCode": 307
      },
      {
        "name": "OpenAI Community",
        "base_url": "https://community.openai.com/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "OMG.lol",
        "base_url": "https://{}.omg.lol",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Polar",
        "base_url": "https://polar.sh/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Quizlet",
        "base_url": "https://quizlet.com/user/{}/sets",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "PyPi",
        "base_url": "https://pypi.org/user/{}",
        "url_probe":"https://pypi.org/_includes/administer-user-include/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Trusted Tutors",
        "base_url": "https://trusted-tutors.co.uk/instructor/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Instructables",
        "base_url": "https://www.instructables.com/member/{}",
        "url_probe": "https://www.instructables.com/json-api/showAuthorExists?screenName={}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Intigriti",
        "base_url": "https://app.intigriti.com/profile/{}",
        "url_probe": "https://api.intigriti.com/user/public/profile/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Ionic Forum",
        "base_url": "https://forum.ionicframework.com/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Issuu",
        "base_url": "https://issuu.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Itch.io",
        "base_url": "https://{}.itch.io/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Itemfix",
        "base_url": "https://www.itemfix.com/c/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<title>ItemFix - Channel: </title>"
      },
      {
        "name": "Jellyfin Weblate",
        "base_url": "https://translate.jellyfin.org/user/{}/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Jimdo",
        "base_url": "https://{}.jimdosite.com",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "Not Found"
      },
      {
        "name": "Joplin Forum",
        "base_url": "https://discourse.joplinapp.org/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Kaggle",
        "base_url": "https://www.kaggle.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Kaskus",
        "base_url": "https://www.kaskus.co.id/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Kick",
        "base_url": "https://kick.com/{}",
        "url_probe": "https://kick.com/api/v2/channels/{}",
        "follow_redirects": true,
        "errorType": "unknown"
      },
      {
        "name": "Kongregate",
        "base_url": "https://www.kongregate.com/accounts/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "LOR",
        "base_url": "https://www.linux.org.ru/people/{}/profile",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Launchpad",
        "base_url": "https://launchpad.net/~{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "LessWrong",
        "base_url": "https://www.lesswrong.com/users/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Letterboxd",
        "base_url": "https://letterboxd.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "LibraryThing",
        "base_url": "https://www.librarything.com/profile/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<p>Error: This user doesn't exist</p>"
      },
      {
        "name": "Lichess",
        "base_url": "https://lichess.org/@/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Listed",
        "base_url": "https://listed.to/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "LiveJournal",
        "base_url": "https://{}.livejournal.com",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Lobsters",
        "base_url": "https://lobste.rs/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "LottieFiles",
        "base_url": "https://lottiefiles.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "MMORPG Forum",
        "base_url": "https://forums.mmorpg.com/profile/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Memrise",
        "base_url": "https://www.memrise.com/user/{}/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Minecraft",
        "base_url": "https://api.mojang.com/users/profiles/minecraft/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "MixCloud",
        "base_url": "https://www.mixcloud.com/{}/",
        "url_probe": "https://api.mixcloud.com/{}/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Monkeytype",
        "base_url": "https://monkeytype.com/profile/{}",
        "url_probe": "https://api.monkeytype.com/users/{}/profile",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Motorradfrage",
        "base_url": "https://www.motorradfrage.net/nutzer/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "MyAnimeList",
        "base_url": "https://myanimelist.net/profile/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "MyMiniFactory",
        "base_url": "https://www.myminifactory.com/users/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "MyDramaList",
        "base_url": "https://www.mydramalist.com/profile/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "Sign in - MyDramaList"
      },
      {
        "name": "Myspace",
        "base_url": "https://myspace.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "NICommunityForum",
        "base_url": "https://community.native-instruments.com/profile/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "NationStates Nation",
        "base_url": "https://nationstates.net/nation={}",
        "follow_redirects": true,
        "errorType": "unknown",
        "errorMsg": "<title>NationStates | Not Found</title>"
      },
      {
        "name": "NationStates Region",
        "base_url": "https://nationstates.net/region={}",
        "follow_redirects": true,
        "errorType": "unknown",
        "errorMsg": "<title>NationStates | Not Found</title>"
      },
      {
        "name": "Naver",
        "base_url": "https://blog.naver.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Needrom",
        "base_url": "https://www.needrom.com/author/{}/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Newgrounds",
        "base_url": "https://{}.newgrounds.com",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Nextcloud Forum",
        "base_url": "https://help.nextcloud.com/u/{}/summary",
        "follow_redirects": true,
        "errorType": "unknown"
      },
      {
        "name": "Nightbot",
        "base_url": "https://nightbot.tv/t/{}/commands",
        "url_probe": "https://api.nightbot.tv/1/channels/t/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "NintendoLife",
        "base_url": "https://www.nintendolife.com/users/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "NitroType",
        "base_url": "https://www.nitrotype.com/racer/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<title>Nitro Type | Competitive Typing Game | Race Your Friends</title>"
      },
      {
        "name": "NotABug.org",
        "base_url": "https://notabug.org/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Nyaa.si",
        "base_url": "https://nyaa.si/user/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "OpenStreetMap",
        "base_url": "https://www.openstreetmap.org/user/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Opensource",
        "base_url": "https://opensource.com/users/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "OurDJTalk",
        "base_url": "https://ourdjtalk.com/members?username={}",
        "follow_redirects": false,
        "errorType": "status_code",
        "errorCode": 301
      },
      {
        "name": "PCGamer",
        "base_url": "https://forums.pcgamer.com/members/?username={}",
        "follow_redirects": false,
        "errorType": "status_code",
        "errorCode": 200
      },
      {
        "name": "PSNProfiles Forum",
        "base_url": "https://forum.psnprofiles.com/profile/{}",
        "follow_redirects": true,
        "errorType": "unknown"
      },
      {
        "name": "Packagist",
        "base_url": "https://packagist.org/packages/{}/",
        "follow_redirects": true,
        "errorType": "response_url",
        "response_url": "https://packagist.org/search/?q={}&reason=vendor_not_found"
      },
      {
        "name": "Pastebin",
        "base_url": "https://pastebin.com/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "PentesterLab",
        "base_url": "https://pentesterlab.com/profile/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "PepperIT",
        "base_url": "https://www.pepper.it/profile/{}/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Periscope",
        "base_url": "https://www.periscope.tv/{}/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Pinkbike",
        "base_url": "https://www.pinkbike.com/u/{}/",
        "follow_redirects": true,
        "errorType": "unknown"
      },
      {
        "name": "Pokemon Showdown",
        "base_url": "https://pokemonshowdown.com/users/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Polarsteps",
        "base_url": "https://polarsteps.com/{}",
        "url_probe": "https://api.polarsteps.com/users/byusername/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Polymart",
        "base_url": "https://polymart.org/user/{}",
        "follow_redirects": true,
        "errorType": "unknown"
      },
      {
        "name": "PromoDJ",
        "base_url": "http://promodj.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Rajce.net",
        "base_url": "https://{}.rajce.idnes.cz/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Rarible",
        "base_url": "https://rarible.com/{}",
        "url_probe": "https://rarible.com/marketplace/api/v4/urls/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Rate Your Music",
        "base_url": "https://rateyourmusic.com/~{}",
        "follow_redirects": true,
        "errorType": "unknown"
      },
      {
        "name": "Rclone Forum",
        "base_url": "https://forum.rclone.org/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Redbubble",
        "base_url": "https://www.redbubble.com/people/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Reisefrage",
        "base_url": "https://www.reisefrage.net/nutzer/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Replit.com",
        "base_url": "https://replit.com/@{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "ResearchGate",
        "base_url": "https://www.researchgate.net/profile/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "ReverbNation",
        "base_url": "https://www.reverbnation.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Roblox",
        "base_url": "https://www.roblox.com/user.aspx?username={}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "RubyGems",
        "base_url": "https://rubygems.org/profiles/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "RuneScape",
        "base_url": "https://apps.runescape.com/runemetrics/app/overview/player/{}",
        "url_probe": "https://apps.runescape.com/runemetrics/profile/profile?user={}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "{\"error\":\"NO_PROFILE\",\"loggedIn\":\"false\"}"
      },
      {
        "name": "SWAPD",
        "base_url": "https://swapd.co/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Sbazar.cz",
        "base_url": "https://www.sbazar.cz/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Scratch",
        "base_url": "https://scratch.mit.edu/users/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Scribd",
        "base_url": "https://www.scribd.com/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "Page not found"
      },
      {
        "name": "ShitpostBot5000",
        "base_url": "https://www.shitpostbot.com/user/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Signal",
        "base_url": "https://community.signalusers.org/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Sketchfab",
        "base_url": "https://sketchfab.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Slack",
        "base_url": "https://{}.slack.com",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Slant",
        "base_url": "https://www.slant.co/users/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Slashdot",
        "base_url": "https://slashdot.org/~{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "user you requested does not exist"
      },
      {
        "name": "SlideShare",
        "base_url": "https://slideshare.net/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<title>Username available</title>"
      },
      {
        "name": "Slides",
        "base_url": "https://slides.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "SmugMug",
        "base_url": "https://{}.smugmug.com",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Smule",
        "base_url": "https://www.smule.com/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "Smule | Page Not Found (404)"
      },
      {
        "name": "SoylentNews",
        "base_url": "https://soylentnews.org/~{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "The user you requested does not exist, no matter how much you wish this might be the case."
      },
      {
        "name": "Speedrun.com",
        "base_url": "https://speedrun.com/users/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Spells8",
        "base_url": "https://forum.spells8.com/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Splits.io",
        "base_url": "https://splits.io/users/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Sporcle",
        "base_url": "https://www.sporcle.com/user/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Sportlerfrage",
        "base_url": "https://www.sportlerfrage.net/nutzer/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "SportsRU",
        "base_url": "https://www.sports.ru/profile/{}/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Star Citizen",
        "base_url": "https://robertsspaceindustries.com/citizens/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Steam Community (Group)",
        "base_url": "https://steamcommunity.com/groups/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "No group could be retrieved for the given URL"
      },
      {
        "name": "SublimeForum",
        "base_url": "https://forum.sublimetext.com/u/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "TETR.IO",
        "base_url": "https://ch.tetr.io/u/{}",
        "url_probe": "https://ch.tetr.io/api/users/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Tiendanube",
        "base_url": "https://{}.mitiendanube.com/",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Topcoder",
        "base_url": "https://profiles.topcoder.com/{}/",
        "url_probe": "https://api.topcoder.com/v5/members/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "TRAKTRAIN",
        "base_url": "https://traktrain.com/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      },
      {
        "name": "Monzo Bank",
        "base_url": "https://monzo.me/{}",
        "follow_redirects": true,
        "errorType": "errorMsg",
        "errorMsg": "<title>Monzo.me – Send money instantly through a link</title>"
      },
      {
        "name": "Modrinth",
        "base_url": "https://modrinth.com/user/{}",
        "follow_redirects": true,
        "errorType": "status_code"
      }
]`

// Глобальная переменная для хранения данных сайтов
var sites []SiteInfo
var once sync.Once // Для однократной загрузки data.json

// Функция для загрузки данных о сайтах
func loadSites() {
	once.Do(func() {
		// Читаем из строковой константы
		if err := json.Unmarshal([]byte(sitesDataStr), &sites); err != nil {
			log.Fatalf("Error unmarshalling embedded sites data: %v", err)
		}
		if len(sites) == 0 {
			log.Fatalf("Embedded sites data seems empty or invalid")
		}
		log.Printf("Loaded %d sites from embedded data", len(sites))
	})
}

// Структура для ответа API
type SearchResult struct {
	Username          string   `json:"username"`
	FoundOn           []string `json:"found_on"`            // Сайты, где найден пользователь
	Breaches          []string `json:"breaches"`            // Найденные утечки (пока не используется)
	Error             string   `json:"error,omitempty"`     // Сообщение об ошибке
	TotalSitesChecked int      `json:"total_sites_checked"` // Общее количество проверенных сайтов
}

// Функция проверки одного сайта
func checkSite(ctx context.Context, client *http.Client, site SiteInfo, username string, resultsChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	checkURL := site.BaseURL
	if site.URLProbe != "" {
		checkURL = site.URLProbe // Используем URL для проверки, если он указан
	}
	targetURL := strings.Replace(checkURL, "{}", username, 1)

	// Добавляем User-Agent, чтобы имитировать браузер
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		// Не логируем ошибку создания запроса, т.к. их может быть много
		// log.Printf("Error creating request for %s: %v", site.Name, err)
		return
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		// Не логируем ошибки сети, т.к. их может быть много
		// log.Printf("Error checking %s (%s): %v", site.Name, targetURL, err)
		return
	}
	defer resp.Body.Close()

	// --- Логика проверки ---
	found := false
	switch site.ErrorType {
	case "status_code":
		// Ожидаем, что errorCode - это число (статус код ошибки)
		var expectedErrorCode int
		switch v := site.ErrorCode.(type) {
		case float64: // JSON числа часто парсятся как float64
			expectedErrorCode = int(v)
		case int:
			expectedErrorCode = v
		default:
			// Не можем обработать - пропускаем
			return
		}
		// Пользователь найден, если статус НЕ равен коду ошибки
		if resp.StatusCode != expectedErrorCode {
			found = true
		}
	case "errorMsg":
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return // Не можем прочитать тело - пропускаем
		}
		bodyString := string(bodyBytes)
		// Пользователь найден, если тело НЕ содержит сообщение об ошибке
		if !strings.Contains(bodyString, site.ErrorMsg) {
			found = true
		}
	case "profilePresence":
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return // Не можем прочитать тело - пропускаем
		}
		bodyString := string(bodyBytes)
		// Пользователь найден, если тело СОДЕРЖИТ сообщение о наличии профиля
		if strings.Contains(bodyString, site.ErrorMsg) {
			found = true
		}
	case "unknown":
		// Не можем определить - пропускаем сайт
		return
	default:
		// Неизвестный тип ошибки - пропускаем
		return
	}

	if found {
		select {
		case resultsChan <- site.Name: // Отправляем имя сайта, если нашли
		case <-ctx.Done(): // Прекращаем, если контекст завершен (например, таймаут)
			return
		}
	}
}

// Handler is the main entry point for Vercel serverless function
func Handler(w http.ResponseWriter, r *http.Request) {
	// Загружаем данные сайтов при первом вызове
	loadSites()

	// Log the incoming request path and method for debugging
	log.Printf("Received request: Method=%s, Path=%s, URL=%s", r.Method, r.URL.Path, r.URL.String())

	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle search endpoint
	if r.URL.Path == "/search" {
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "Username parameter is required", http.StatusBadRequest)
			return
		}

		// Get Telegram API token from environment
		token := os.Getenv("TELEGRAM_BOT_TOKEN")
		if token == "" {
			log.Println("Error: TELEGRAM_BOT_TOKEN environment variable not set")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		// Create HTTP client with timeout
		// Увеличим общий таймаут, т.к. проверяем много сайтов
		client := &http.Client{
			Timeout: 35 * time.Second, // Общий таймаут для всех запросов
		}

		// --- Проверка Telegram ---
		var wgTelegram sync.WaitGroup
		telegramFoundChan := make(chan bool, 1)
		wgTelegram.Add(1)
		go func() {
			defer wgTelegram.Done()
			chatURL := fmt.Sprintf("https://api.telegram.org/bot%s/getChat?chat_id=@%s", token, username)
			reqTg, _ := http.NewRequest("GET", chatURL, nil)  // Создаем запрос для добавления User-Agent
			reqTg.Header.Set("User-Agent", "GoSearchBot/1.0") // Добавляем User-Agent
			chatResp, err := client.Do(reqTg)
			if err != nil {
				log.Printf("Error making request to Telegram API getChat: %v", err)
				telegramFoundChan <- false
				return
			}
			defer chatResp.Body.Close()

			chatBody, err := io.ReadAll(chatResp.Body)
			if err != nil {
				log.Printf("Error reading Telegram API getChat response: %v", err)
				telegramFoundChan <- false
				return
			}

			var chatResult map[string]interface{}
			if err := json.Unmarshal(chatBody, &chatResult); err != nil {
				log.Printf("Error parsing Telegram API getChat response: %v", err)
				telegramFoundChan <- false
				return
			}

			if okValue, okType := chatResult["ok"].(bool); okType && okValue {
				telegramFoundChan <- true
				return
			}
			telegramFoundChan <- false
		}()

		// --- Проверка утечек паролей через Have I Been Pwned API ---
		var wgBreaches sync.WaitGroup
		breachesChan := make(chan []string, 1) // Канал для результата проверки утечек
		wgBreaches.Add(1)
		go func() {
			defer wgBreaches.Done()
			breachesFound := checkPasswordBreaches(client, username)
			breachesChan <- breachesFound
		}()

		// --- Проверка сайтов из data.json ---
		var wgSites sync.WaitGroup
		resultsChan := make(chan string, len(sites)) // Канал для имен найденных сайтов
		// Контекст с таймаутом для всех проверок сайтов
		ctx, cancel := context.WithTimeout(context.Background(), 38*time.Second) // Увеличиваем таймаут для сайтов
		defer cancel()                                                           // Важно отменить контекст

		sitesCount := len(sites) // Общее количество сайтов для проверки
		log.Printf("Starting to check %d sites for username: %s", sitesCount, username)

		for _, site := range sites {
			wgSites.Add(1)
			go checkSite(ctx, client, site, username, resultsChan, &wgSites)
		}

		// Горутина для ожидания завершения всех проверок сайтов
		go func() {
			wgSites.Wait()
			close(resultsChan) // Закрываем канал, когда все горутины завершились
		}()

		// Сбор результатов
		foundSites := []string{}

		// Ждем результат от Telegram
		wgTelegram.Wait()
		if <-telegramFoundChan {
			foundSites = append(foundSites, "Telegram")
		}

		// Собираем результаты от проверки сайтов
		for siteName := range resultsChan {
			foundSites = append(foundSites, siteName)
		}

		// Ждем результаты проверки утечек
		wgBreaches.Wait()
		breachesResult := <-breachesChan

		// Формируем финальный ответ
		finalResult := SearchResult{
			Username:          username,
			FoundOn:           foundSites,
			Breaches:          breachesResult,
			TotalSitesChecked: sitesCount + 1, // +1 за Telegram
		}

		if len(foundSites) == 0 {
			finalResult.Error = "Пользователь не найден ни на одном из проверяемых сайтов."
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(finalResult)
		return
	}

	// Handle root endpoint
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("GoSearch API (Vercel) with multi-site check - Simplified Version"))
		return
	}

	// Handle 404 for any other paths
	http.Error(w, "Not Found", http.StatusNotFound)
}

// Функция для проверки утечек паролей через Have I Been Pwned API
func checkPasswordBreaches(client *http.Client, username string) []string {
	breaches := []string{}

	// Проверяем наличие e-mail в утечках
	if strings.Contains(username, "@") {
		// Проверка e-mail на Have I Been Pwned
		emailUrl := fmt.Sprintf("https://haveibeenpwned.com/api/v3/breachedaccount/%s", username)
		req, err := http.NewRequest("GET", emailUrl, nil)
		if err != nil {
			log.Printf("Error creating request to HIBP: %v", err)
			return breaches
		}

		// Добавляем необходимые заголовки для API
		req.Header.Set("User-Agent", "GoSearchBot/1.0")
		req.Header.Set("hibp-api-key", os.Getenv("HIBP_API_KEY")) // Лучше использовать API ключ

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error checking HIBP: %v", err)
			return breaches
		}
		defer resp.Body.Close()

		// Если нашли утечки (код 200)
		if resp.StatusCode == 200 {
			// Парсим JSON ответ
			var breachData []map[string]interface{}
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading HIBP response: %v", err)
				return breaches
			}

			if err := json.Unmarshal(bodyBytes, &breachData); err != nil {
				log.Printf("Error parsing HIBP response: %v", err)
				return breaches
			}

			// Извлекаем названия утечек
			for _, breach := range breachData {
				if name, ok := breach["Name"].(string); ok {
					breaches = append(breaches, name)
				}
			}
		}
	}

	// Если это похоже на логин/никнейм, пробуем альтернативные варианты
	// Это простая эвристика, в реальном сценарии нужен более сложный алгоритм
	if !strings.Contains(username, "@") && len(username) >= 4 {
		// Проверяем общие комбинации с пользовательским именем
		commonDomains := []string{"gmail.com", "yahoo.com", "hotmail.com", "outlook.com"}
		for _, domain := range commonDomains {
			testEmail := username + "@" + domain

			// Пытаемся найти этот вариант email
			emailUrl := fmt.Sprintf("https://haveibeenpwned.com/api/v3/breachedaccount/%s", testEmail)
			req, err := http.NewRequest("GET", emailUrl, nil)
			if err != nil {
				continue
			}

			req.Header.Set("User-Agent", "GoSearchBot/1.0")
			req.Header.Set("hibp-api-key", os.Getenv("HIBP_API_KEY"))

			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			// Если нашли утечки
			if resp.StatusCode == 200 {
				var breachData []map[string]interface{}
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					continue
				}

				if err := json.Unmarshal(bodyBytes, &breachData); err != nil {
					continue
				}

				// Добавляем уникальные утечки
				for _, breach := range breachData {
					if name, ok := breach["Name"].(string); ok {
						// Проверяем, не добавлена ли уже эта утечка
						alreadyAdded := false
						for _, existingBreach := range breaches {
							if existingBreach == name {
								alreadyAdded = true
								break
							}
						}

						if !alreadyAdded {
							breaches = append(breaches, name+" (возможно "+testEmail+")")
						}
					}
				}
			}

			// Делаем паузу, чтобы не превысить лимиты API
			time.Sleep(1500 * time.Millisecond)
		}
	}

	return breaches
}
