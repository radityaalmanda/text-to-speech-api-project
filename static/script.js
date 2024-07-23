document.getElementById('translate-button').addEventListener('click', function(event) {
    const textInput = document.getElementById('text-input').value;
    const sourceLanguage = document.getElementById('source-language-select').value;
    const targetLanguage = document.getElementById('target-language-select').value;

    fetch('/translate', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
            text: textInput,
            sourceLanguage: sourceLanguage,
            targetLanguage: targetLanguage
        })
    })
        .then(response => response.json())
        .then(data => {
            document.getElementById('translated-text').value = data.translatedText;
            resetAudioPlayer();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('An error occurred while translating the text. Please try again.');
        });
});

document.getElementById('translator-form').addEventListener('submit', function(event) {
    event.preventDefault();
    const textInput = document.getElementById('translated-text').value;
    const targetLanguage = document.getElementById('target-language-select').value;

    fetch('/synthesize', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
            text: textInput,
            language: targetLanguage
        })
    })
        .then(response => response.json())
        .then(data => {
            const audioPlayer = document.getElementById('audio-player');
            const audioSource = document.getElementById('audio-source');
            audioSource.src = data.audioPath;
            audioPlayer.style.display = 'block';
            audioPlayer.load();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('An error occurred while synthesizing the text. Please try again.');
        });
});

document.getElementById('reset-button').addEventListener('click', function() {
    document.getElementById('text-input').value = '';
    document.getElementById('translated-text').value = '';
    resetAudioPlayer();
});

function resetAudioPlayer() {
    const audioPlayer = document.getElementById('audio-player');
    audioPlayer.style.display = 'none';
    audioPlayer.pause();
    audioPlayer.currentTime = 0;
}
document.getElementById("translator-form").addEventListener("submit", async function(event) {
    event.preventDefault();

    const text = document.getElementById("text").value;
    const lang = document.getElementById("lang").value;

    const response = await fetch("/translate", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({ text, lang })
    });

    if (response.ok) {
        const data = await response.json();
        const audioPlayer = document.getElementById("audio-player");
        const audioSource = document.getElementById("audio-source");

        audioSource.src = data.audioUrl;
        audioPlayer.style.display = "block";
        audioPlayer.load();
        audioPlayer.play();
    }
});
