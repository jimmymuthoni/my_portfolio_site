//this is my portfolio build from scretch using golang rendering HTML files
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Skill struct {
	Name  string 
	Group string 
}

type Project struct {
	Name        string
	Description string
	Tag         string 
	Stars       int
	Forks       int
	URL         string
	Featured    bool 
}

// SocialLink is one entry in the contact footer.
type SocialLink struct {
	Label string
	Value string
	URL   string
	Icon  string 
}

// PageData is everything the templates need to render the page.
type PageData struct {
	Name        string
	Handle      string
	Tagline     string
	Location    string
	Blurb       string
	PhotoURL    string
	ResumeURL   string
	Skills      []Skill
	SkillGroups []string
	Projects    []Project
	Socials     []SocialLink
	GitHubStats GitHubStats
	Year        string
}

// GitHubStats powers the small counters in the hero panel.
type GitHubStats struct {
	Repos     int
	Stars     int
	Followers int
}


// Live GitHub stats.
const githubUsername = "jimmymuthoni"

type repoCount struct {
	Stars int
	Forks int
}

type liveStats struct {
	Totals GitHubStats
	Repos  map[string]repoCount 
}

var (
	statsMu    sync.RWMutex
	statsCache = liveStats{
		// Fallback values, used only if the very first fetch on
		// startup fails before any real data has been cached.
		Totals: GitHubStats{Repos: 22, Stars: 18, Followers: 20},
		Repos:  map[string]repoCount{},
	}
	statsFetchedAt time.Time
)

type ghUser struct {
	PublicRepos int `json:"public_repos"`
	Followers   int `json:"followers"`
}

type ghRepo struct {
	Name            string `json:"name"`
	Fork            bool   `json:"fork"`
	StargazersCount int    `json:"stargazers_count"`
	ForksCount      int    `json:"forks_count"`
}

func fetchLiveStats() (liveStats, error) {
	client := &http.Client{Timeout: 8 * time.Second}

	user, err := fetchJSON[ghUser](client, "https://api.github.com/users/"+githubUsername)
	if err != nil {
		return liveStats{}, fmt.Errorf("fetch user: %w", err)
	}

	repos, err := fetchJSON[[]ghRepo](client, "https://api.github.com/users/"+githubUsername+"/repos?per_page=100")
	if err != nil {
		return liveStats{}, fmt.Errorf("fetch repos: %w", err)
	}

	out := liveStats{Repos: map[string]repoCount{}}
	out.Totals.Repos = user.PublicRepos
	out.Totals.Followers = user.Followers

	starTotal := 0
	for _, repo := range repos {
		out.Repos[repo.Name] = repoCount{Stars: repo.StargazersCount, Forks: repo.ForksCount}
		if !repo.Fork {
			starTotal += repo.StargazersCount
		}
	}
	out.Totals.Stars = starTotal

	return out, nil
}

func fetchJSON[T any](client *http.Client, url string) (T, error) {
	var zero T
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return zero, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "jimmymuthoni-portfolio")

	resp, err := client.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return zero, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	var out T
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return zero, err
	}
	return out, nil
}

func refreshLiveStats() {
	fresh, err := fetchLiveStats()
	if err != nil {
		log.Printf("github stats refresh failed, keeping last known values: %v", err)
		return
	}
	statsMu.Lock()
	statsCache = fresh
	statsFetchedAt = time.Now()
	statsMu.Unlock()
	log.Printf("github stats refreshed: %+v", fresh.Totals)
}

func getLiveStats() liveStats {
	statsMu.RLock()
	defer statsMu.RUnlock()
	return statsCache
}

// startStatsRefresher does one blocking fetch (so the very first page
// load already has real numbers), then keeps refreshing in the
// background on the given interval for as long as the process runs.
func startStatsRefresher(interval time.Duration) {
	refreshLiveStats()
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			refreshLiveStats()
		}
	}()
}

func siteData() PageData {
	live := getLiveStats()

	data := PageData{
		Name:      "James Kamau",
		Handle:    "Jay",
		Tagline:   "Software · Cloud · DevOps · MLOps · SRE",
		Location:  "Nairobi, Kenya — UTC+03:00",
		Blurb:     "Software and cloud engineer with strong Computer Science background focused on building backend systems, cloud infrastructure, and the occasional AI system that has to earn its place in production. Based in Nairobi,Kenya",
		PhotoURL:  "/static/images/profile.jpg",
		ResumeURL: "/static/resume.pdf",
		Skills: []Skill{
			{"Go", "Languages"},
			{"Python", "Languages"},
			{"C++", "Languages"},
			{"Java", "Languages"},
			{"Bash", "Languages"},
			{"Fiber", "Frameworks"},
			{"FastAPI", "Frameworks"},
			{"Django", "Frameworks"},
			{"PostgreSQL", "Data"},
			{"MongoDB", "Data"},
			{"MySQL", "Data"},
			{"AWS", "Cloud & Ops"},
			{"Docker", "Cloud & Ops"},
			{"Kubernetes", "Cloud & Ops"},
			{"Terraform / IaC", "Cloud & Ops"},
			{"Linux", "Cloud & Ops"},
			{"MLOps", "Cloud & Ops"},
		},
		SkillGroups: []string{"Languages", "Frameworks", "Data", "Cloud & Ops"},
		Projects: []Project{
			{
				Name:        "queue_forge",
				Description: "Cloud-native asynchronous report generation API built with Go, PostgreSQL, Amazon SQS, and S3.",
				Tag:         "Go",
				Stars:       4,
				Forks:       0,
				URL:         "https://github.com/jimmymuthoni/queue_forge",
				Featured:    true,
			},
			{
				Name:        "network_security_system",
				Description: "ML-powered cybersecurity tool that identifies and blocks phishing websites in real time.",
				Tag:         "Python",
				Stars:       13,
				Forks:       5,
				URL:         "https://github.com/jimmymuthoni/network_security_system",
				Featured:    true,
			},
			{
				Name:        "customers_support_system",
				Description: "RAG-powered e-commerce support and recommendation bot trained on the Flipkart product dataset.",
				Tag:         "Python",
				Stars:       10,
				Forks:       0,
				URL:         "https://github.com/jimmymuthoni/customers_support_system",
			},
			{
				Name:        "gpt_inference_engine",
				Description: "A high-performance C/C++ inference engine that runs trained models efficiently on CPU.",
				Tag:         "C++",
				Stars:       4,
				Forks:       1,
				URL:         "https://github.com/jimmymuthoni/gpt_inference_engine",
			},
			{
				Name:        "grandmaster",
				Description: "A Qt6 chess game in C++ with full move validation, highlighting, and sound effects.",
				Tag:         "C++",
				Stars:       11,
				Forks:       3,
				URL:         "https://github.com/jimmymuthoni/grandmaster",
			},
		},
		Socials: []SocialLink{
			{Label: "GitHub", Value: "jimmymuthoni", URL: "https://github.com/jimmymuthoni", Icon: "github"},
			{Label: "LinkedIn", Value: "jimmy-muthoni", URL: "https://www.linkedin.com/in/james-kamau-356636303/", Icon: "linkedin"},
			{Label: "Email", Value: "jimkamau001@gmail.com", URL: "mailto:jimkamau001@gmail.com", Icon: "mail"},
			{Label: "Instagram", Value: "its_digital_nomad", URL: "https://www.instagram.com/its_digital_nomad/", Icon: "instagram"},
		
		},
		GitHubStats: live.Totals,
		Year:        "2026",
	}

	// Fill in real star/fork counts on the project cards, keyed by
	// repo name, so those never need manual updates either.
	for i, p := range data.Projects {
		name := strings.TrimSpace(strings.TrimPrefix(p.URL, "https://github.com/"+githubUsername+"/"))
		if counts, ok := live.Repos[name]; ok {
			data.Projects[i].Stars = counts.Stars
			data.Projects[i].Forks = counts.Forks
		}
	}

	return data
}

var tmpl *template.Template

func loadTemplates() *template.Template {
	return template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/index.html",
	))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", siteData()); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func main() {
	tmpl = loadTemplates()
	startStatsRefresher(time.Hour)

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/healthz", healthHandler)

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
