// global variables
var songVolume = 1.0;

function formatTime(totalSec) {
  var minutes = Math.floor(totalSec / 60);
  var seconds = Math.floor(totalSec % 60);
  return `${minutes}:${seconds < 10 ? "0" : ""}${seconds}`;
}

function onAudioLoadMetadata(audio) {
  const progress = document.getElementById("progress");
  progress.min = 0;
  progress.max = audio.duration;
  progress.value = audio.currentTime;

  duration.innerHTML = `${formatTime(audio.duration ? audio.duration : 0)}`;
  currentTime.innerHTML = `${formatTime(audio.currentTime)}`;
}

function playSong(musicPath) {
  const audio = document.getElementById("audio");
  const source = document.getElementById("source");

  if (musicPath) {
    // Stop current playback
    source.src = `/play?music-path=${encodeURIComponent(musicPath)}`;
    audio.load();
    document.getElementById("player").style.display = "flex";
  }

  // Wait for the audio to be ready before playing
  const playWhenReady = () => {
    audio
      .play()
      .then(() => {
        console.log("Playback started");
      })
      .catch((err) => {
        console.error("Playback failed:", err);
      });

    document.querySelector('button.control-btn[title="Play/Pause"]').innerHTML =
      `<svg xmlns="http://www.w3.org/2000/svg" width="50" height="50" fill="white" viewBox="0 0 24 24">
         <circle cx="12" cy="12" r="10" fill="white"/>
         <rect x="9" y="8" width="2" height="8" fill="black"/>
         <rect x="13" y="8" width="2" height="8" fill="black"/>
      </svg>`;

    // Remove the event listener after playing
    audio.removeEventListener("canplay", playWhenReady);
  };

  // Only call play if the audio is not ready
  if (audio.readyState < 3) {
    // HAVE_FUTURE_DATA or higher
    audio.addEventListener("canplay", playWhenReady);
  } else {
    playWhenReady();
  }
}

function playSongFromJsonResponce(data) {
  let nextSongId = data["id"];
  let nextSongPath = data["path"];

  fetch(`/song/details?id=${nextSongId}&toPlay=true`)
    .then((response) => response.text())
    .then((html) => {
      document.getElementById("music-details").innerHTML = html;
      playSong(nextSongPath);
    })
    .catch((err) => {
      console.error("ERROR: fetching:", err);
    });
}

function playAll(type, value) {
  fetch(`/play-all?type=${type}&value=${encodeURIComponent(value)}`)
    .then((response) => response.json())
    .then((data) => playSongFromJsonResponce(data))
    .catch((err) => {
      console.error("ERROR: fetching: ", err);
    });
}

function playNextSong() {
  fetch("/next-song")
    .then((response) => response.json())
    .then((data) => playSongFromJsonResponce(data))
    .catch((err) => {
      console.error("ERROR: fetching:", err);
    });
}

function playPreviousSong() {
  fetch("/previous-song")
    .then((response) => response.json())
    .then((data) => {
      let prevSongId = data["id"];
      let prevSongPath = data["path"];

      fetch(`/song/details?id=${prevSongId}&toPlay=true`)
        .then((response) => response.text())
        .then((html) => {
          document.getElementById("music-details").innerHTML = html;
          playSong(prevSongPath);
        })
        .catch((err) => {
          console.error("ERROR: fetching:", err);
        });
    })
    .catch((err) => {
      console.error("ERROR: fetching: ", err);
    });
}

function togglePlay() {
  const audio = document.getElementById("audio");
  const playPauseBtn = document.querySelector(
    'button.control-btn[title="Play/Pause"]',
  );

  if (audio.readyState < 3) {
    console.error("ERROR: no song is not loaded to player");
    return;
  }

  if (audio.paused) {
    playSong();
  } else {
    audio.pause();
    playPauseBtn.innerHTML = `
      <svg xmlns="http://www.w3.org/2000/svg" width="50" height="50" fill="white" viewBox="0 0 24 24">
        <circle cx="12" cy="12" r="10" fill="white" />
        <polygon points="10,8 16,12 10,16" fill="black" />
      </svg>`;
    console.log("audio is paused");
  }
}

function toggleLoop() {
  const audio = document.getElementById("audio");
  const loopBtn = document.querySelector('button.control-btn[title="Loop"]');

  audio.loop = !audio.loop;
  if (audio.loop) {
    loopBtn.innerHTML = `
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="-1 -1 26 26"  width="24" height="24" fill="none" stroke="gray" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" >
        <polyline points="17 1 21 5 17 9"></polyline>
        <path d="M3 11V9a4 4 0 014-4h14"></path>
        <polyline points="7 23 3 19 7 15"></polyline>
        <path d="M21 13v2a4 4 0 01-4 4H3"></path>
        <circle cx="2" cy="2" r="1.5" fill="gray" />
      </svg>
      `;
  } else {
    loopBtn.innerHTML = `
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="-1 -1 26 26"  width="24" height="24" fill="none" stroke="gray" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" >
        <polyline points="17 1 21 5 17 9"></polyline>
        <path d="M3 11V9a4 4 0 014-4h14"></path>
        <polyline points="7 23 3 19 7 15"></polyline>
        <path d="M21 13v2a4 4 0 01-4 4H3"></path>
      </svg>`;
  }
}

function adjustVolume(value) {
  const audio = document.getElementById("audio");

  document.getElementById("volume-value").textContent = value + "%";
  songVolume = parseInt(value) / 100;
  audio.volume = songVolume;
}

function toggleMute() {
  const audio = document.getElementById("audio");
  const mute = !audio.muted;
  audio.muted = mute;

  const icon = document.querySelector("#mute-toggle-lable .mute-toggle-icon");
  const volumeValue = document.getElementById("volume-value");
  const volumeBar = document.getElementById("volume-bar");

  if (mute) {
    volumeValue.textContent = "0%";
    volumeBar.value = 0;
    icon.innerHTML = `
      <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="none" stroke="gray" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24" >
          <path d="M11 5L6 9H2v6h4l5 4V5z"></path>
          <line x1="23" y1="9" x2="17" y2="15"></line>
          <line x1="17" y1="9" x2="23" y2="15"></line>
      </svg>`;
  } else {
    volumeValue.textContent = songVolume * 100 + "%";
    volumeBar.value = songVolume * 100;
    icon.innerHTML = `
      <!-- unmute -->
      <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="none" stroke="gray" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24" >
          <path d="M11 5L6 9H2v6h4l5 4V5z"></path>
          <path d="M15.54 8.46a5 5 0 010 7.07"></path>
          <path d="M19.07 4.93a9 9 0 010 12.73"></path>
      </svg>`;
  }
}
