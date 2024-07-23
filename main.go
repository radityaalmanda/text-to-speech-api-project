package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
)

var (
	translateClient    *translate.Client
	ttsClient          *texttospeech.Client
	ttsMutex           sync.Mutex // Mutex to protect TTS client initialization
	initOnce           sync.Once
	supportedLanguages = map[string]string{
		"id":    "id",
		"en":    "en",
		"ja":    "ja",
		"zh":    "zh",
		"ar":    "ar",
		"fr":    "fr",
		"es":    "es",
		"de":    "de",
		"ru":    "ru",
		"ko":    "ko",
		"it":    "it",
		"pt":    "pt",
		"nl":    "nl",
		"sv":    "sv",
		"no":    "no",
		"da":    "da",
		"fi":    "fi",
		"cs":    "cs",
		"tr":    "tr",
		"he":    "he",
		"el":    "el",
		"hi":    "hi",
		"th":    "th",
		"vi":    "vi",
		"id-ID": "id-ID",
		"en-US": "en-US",
		"ja-JP": "ja-JP",
		"zh-CN": "zh-CN",
		"ar-XA": "ar-XA",
		"fr-FR": "fr-FR",
		"es-ES": "es-ES",
		"de-DE": "de-DE",
		"ru-RU": "ru-RU",
		"ko-KR": "ko-KR",
		"it-IT": "it-IT",
		"pt-BR": "pt-BR",
		"nl-NL": "nl-NL",
		"sv-SE": "sv-SE",
		"no-NO": "no-NO",
		"da-DK": "da-DK",
		"fi-FI": "fi-FI",
		"pl-PL": "pl-PL",
		"cs-CZ": "cs-CZ",
		"tr-TR": "tr-TR",
		"he-IL": "he-IL",
		"el-GR": "el-GR",
		"hi-IN": "hi-IN",
		"th-TH": "th-TH",
		"vi-VN": "vi-VN",
	}
)

func initializeClients() {
	var err error
	translateClient, err = translate.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create translate client: %v", err)
	}

	ttsClient, err = texttospeech.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create TTS client: %v", err)
	}
}

func main() {
	initOnce.Do(initializeClients)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/translate", translateHandler)
	http.HandleFunc("/synthesize", synthesizeHandler)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

func translateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	text := r.FormValue("text")
	sourceLang := r.FormValue("sourceLanguage")
	targetLang := r.FormValue("targetLanguage")
	if text == "" || sourceLang == "" || targetLang == "" {
		http.Error(w, "Text, source language, or target language not provided", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	target, err := language.Parse(supportedLanguages[targetLang])
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid target language: %v", err), http.StatusBadRequest)
		return
	}
	source, err := language.Parse(supportedLanguages[sourceLang])
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid source language: %v", err), http.StatusBadRequest)
		return
	}

	resp, err := translateClient.Translate(ctx, []string{text}, target, &translate.Options{
		Source: source,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Translation error: %v", err), http.StatusInternalServerError)
		return
	}

	translatedText := ""
	if len(resp) > 0 {
		translatedText = resp[0].Text
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"translatedText": "%s"}`, translatedText)
}

func synthesizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	text := r.FormValue("text")
	language := r.FormValue("language")
	if text == "" || language == "" {
		http.Error(w, "Text or language not provided", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	ttsReq := textToSpeechRequest(text, supportedLanguages[language])
	resp, err := ttsClient.SynthesizeSpeech(ctx, ttsReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Synthesize error: %v", err), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("static/output_%d.mp3", time.Now().UnixNano())
	err = ioutil.WriteFile(filename, resp.AudioContent, 0644)
	if err != nil {
		http.Error(w, fmt.Sprintf("Write file error: %v", err), http.StatusInternalServerError)
		return
	}

	audioPath := fmt.Sprintf("/%s", filename)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	fmt.Fprintf(w, `{"audioPath": "%s"}`, audioPath)

}

func initTTSClient(language string) error {
	ttsMutex.Lock()
	defer ttsMutex.Unlock()

	// Close the existing TTS client if it exists
	if ttsClient != nil {
		ttsClient.Close()
	}

	var err error
	ttsClient, err = texttospeech.NewClient(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to create TTS client: %v", err)
	}
	return nil
}

func textToSpeechRequest(text, language string) *texttospeechpb.SynthesizeSpeechRequest {
	return &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: language,
			SsmlGender:   texttospeechpb.SsmlVoiceGender_NEUTRAL,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}
}
