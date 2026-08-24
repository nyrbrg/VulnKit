package labs

type Difficulty string

const (
    Apprentice   Difficulty = "apprentice"
    Practitioner Difficulty = "practitioner"
    Expert       Difficulty = "expert"
)

type Lab struct {
    ID          string     `json:"id"`
    Title       string     `json:"title"`
    Category    string     `json:"category"`
    Difficulty  Difficulty `json:"difficulty"`
    Description string     `json:"description"`
    Objective   string     `json:"objective"`
    Hint        string     `json:"hint"`
    Solution    string     `json:"solution"`
    Path        string     `json:"path"`
}

var DefaultLabs = []Lab{
    {
        ID:          "xss-reflected-basic",
        Title:       "Reflected XSS – nothing encoded",
        Category:    "XSS",
        Difficulty:  Apprentice,
        Description: "The search functionality reflects user input directly into the HTML response without any encoding.",
        Objective:   "Perform a reflected XSS attack that calls the alert() function.",
        Hint:        "Try injecting a script tag into the search parameter.",
        Solution:    "<script>alert(1)</script>",
        Path:        "/xss/reflected.php",
    },
}