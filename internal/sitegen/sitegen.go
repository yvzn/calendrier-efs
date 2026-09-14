// Package sitegen renders the static dist/index.html (French) and
// dist/en/index.html (English) pages that list every saved location's
// subscribable .ics calendar.
package sitegen

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"efs-ical/internal/store"
)

//go:embed templates/page.html.tmpl
var templateFS embed.FS

//go:embed static/style.css static/script.js
var staticFS embed.FS

type locationView struct {
	Name               string
	PostalCode         string
	ICSHref            string
	ICSAbsoluteURL     string // empty when baseURL is unknown at build time
	SubscribeAriaLabel string
	CopyAriaLabel      string
}

type howTo struct {
	Heading string
	Google  string
	Outlook string
	Apple   string
	Android string
}

type pageData struct {
	Lang             string
	Title            string
	ValueProp        string
	Reassurance      []string
	SectionHeading   string
	HowToSubscribe   howTo
	HowToUnsubscribe howTo
	SubscribeLabel   string
	CopyLabel        string
	CopiedLabel      string
	DirectLinkLabel  string
	Locations        []locationView
	AssetPrefix      string
	SelfLangLabel    string
	SwitchLangHref   string
	SwitchLangLabel  string
}

// Generate writes dist/index.html, dist/en/index.html and the shared
// style.css/script.js, from the current set of saved locations. baseURL,
// when non-empty (e.g. "https://user.github.io/repo/"), is used to render
// the absolute subscription URL for every calendar directly in the HTML, so
// the page works without JavaScript. Pass "" for local/dev builds.
func Generate(distDir string, locations []store.Location, baseURL string) error {
	tmpl, err := template.ParseFS(templateFS, "templates/page.html.tmpl")
	if err != nil {
		return fmt.Errorf("parse site template: %w", err)
	}

	if err := writeStatic(distDir); err != nil {
		return err
	}

	enDir := filepath.Join(distDir, "en")
	if err := os.MkdirAll(enDir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", enDir, err)
	}

	fr := frPage(locations, baseURL)
	if err := renderPage(tmpl, filepath.Join(distDir, "index.html"), fr); err != nil {
		return err
	}

	en := enPage(locations, baseURL)
	if err := renderPage(tmpl, filepath.Join(enDir, "index.html"), en); err != nil {
		return err
	}

	return nil
}

func frPage(locations []store.Location, baseURL string) pageData {
	locs := make([]locationView, len(locations))
	for i, l := range locations {
		locs[i] = locationView{
			Name:               l.Name,
			PostalCode:         l.PostalCode,
			ICSHref:            l.Normalized + ".ics",
			ICSAbsoluteURL:     absoluteURL(baseURL, l.Normalized),
			SubscribeAriaLabel: fmt.Sprintf("S'abonner au calendrier — %s (%s)", l.Name, l.PostalCode),
			CopyAriaLabel:      fmt.Sprintf("Copier le lien ICS — %s (%s)", l.Name, l.PostalCode),
		}
	}
	return pageData{
		Lang:      "fr",
		Title:     "Calendriers de collecte de sang EFS",
		ValueProp: "Ajoutez votre ville à votre agenda et recevez automatiquement les mises à jour.",
		Reassurance: []string{
			"Gratuit",
			"Mise à jour automatique",
			"Désabonnement à tout moment",
		},
		SectionHeading: "Calendriers disponibles",
		HowToSubscribe: howTo{
			Heading: "Besoin d'aide pour l'abonnement ?",
			Google:  "Google Agenda : «Autres agendas» puis «À partir d'une URL», collez le lien ci-dessus.",
			Outlook: "Outlook : «Ajouter un calendrier» puis «S'abonner depuis le web», collez le lien ci-dessus.",
			Apple:   "Calendrier (macOS/iOS) : «Fichier» puis «Nouvel abonnement» (ou appuyez sur le lien depuis l'iPhone/iPad), collez le lien ci-dessus.",
			Android: "Android : pas d'abonnement natif intégré ; ouvrez le lien dans Google Agenda (web) ou utilisez une application comme ICSx5.",
		},
		HowToUnsubscribe: howTo{
			Heading: "Comment se désabonner ?",
			Google:  "Google Agenda : ouvrez «Autres agendas», survolez le calendrier puis cliquez sur «Supprimer».",
			Outlook: "Outlook : clic droit sur le calendrier dans la liste de gauche, puis «Supprimer».",
			Apple:   "Calendrier (macOS/iOS) : sélectionnez le calendrier puis «Supprimer l'abonnement».",
			Android: "Android : supprimez le calendrier depuis l'application utilisée pour vous abonner (Google Agenda ou ICSx5).",
		},
		SubscribeLabel:  "S'abonner au calendrier",
		CopyLabel:       "Copier le lien",
		CopiedLabel:     "Copié !",
		DirectLinkLabel: "Adresse complète du calendrier",
		Locations:       locs,
		AssetPrefix:     "",
		SelfLangLabel:   "Français",
		SwitchLangHref:  "en/",
		SwitchLangLabel: "English",
	}
}

func enPage(locations []store.Location, baseURL string) pageData {
	locs := make([]locationView, len(locations))
	for i, l := range locations {
		locs[i] = locationView{
			Name:               l.Name,
			PostalCode:         l.PostalCode,
			ICSHref:            "../" + l.Normalized + ".ics",
			ICSAbsoluteURL:     absoluteURL(baseURL, l.Normalized),
			SubscribeAriaLabel: fmt.Sprintf("Subscribe to calendar — %s (%s)", l.Name, l.PostalCode),
			CopyAriaLabel:      fmt.Sprintf("Copy ICS link — %s (%s)", l.Name, l.PostalCode),
		}
	}
	return pageData{
		Lang:      "en",
		Title:     "EFS Blood Donation Calendars",
		ValueProp: "Add your city to your calendar and get automatic updates.",
		Reassurance: []string{
			"Free",
			"Updates automatically",
			"Unsubscribe anytime",
		},
		SectionHeading: "Available calendars",
		HowToSubscribe: howTo{
			Heading: "Need help subscribing?",
			Google:  "Google Calendar: “Other calendars” then “From URL”, paste the link above.",
			Outlook: "Outlook: “Add calendar” then “Subscribe from web”, paste the link above.",
			Apple:   "Apple Calendar (macOS/iOS): “File” then “New Calendar Subscription” (or tap the link on iPhone/iPad), paste the link above.",
			Android: "Android: no built-in subscription support; open the link in Google Calendar (web) or use an app like ICSx5.",
		},
		HowToUnsubscribe: howTo{
			Heading: "How do I unsubscribe?",
			Google:  "Google Calendar: open “Other calendars”, hover the calendar and click “Remove”.",
			Outlook: "Outlook: right-click the calendar in the left-hand list, then “Delete”.",
			Apple:   "Apple Calendar (macOS/iOS): select the calendar, then “Unsubscribe”.",
			Android: "Android: remove the calendar from whichever app you used to subscribe (Google Calendar or ICSx5).",
		},
		SubscribeLabel:  "Subscribe to calendar",
		CopyLabel:       "Copy link",
		CopiedLabel:     "Copied!",
		DirectLinkLabel: "Full calendar address",
		Locations:       locs,
		AssetPrefix:     "../",
		SelfLangLabel:   "English",
		SwitchLangHref:  "../",
		SwitchLangLabel: "Français",
	}
}

func absoluteURL(baseURL, normalized string) string {
	if baseURL == "" {
		return ""
	}
	return baseURL + normalized + ".ics"
}

func renderPage(tmpl *template.Template, path string, data pageData) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("render %s: %w", path, err)
	}
	return nil
}

func writeStatic(distDir string) error {
	for _, name := range []string{"style.css", "script.js"} {
		data, err := staticFS.ReadFile("static/" + name)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", name, err)
		}
		if err := os.WriteFile(filepath.Join(distDir, name), data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	return nil
}
