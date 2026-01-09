package screenscraper

// Header contains common API response metadata
type Header struct {
	APIVersion       string `json:"APIversion"`
	CommandRequested string `json:"commandRequested"`
	DateTime         string `json:"dateTime"`
	Error            string `json:"error"`
	Success          string `json:"success"`
}

// ServerInfo contains server infrastructure information (included in most responses).
// CPU values are averages of the last 5 minutes. Thread values indicate simultaneous API access limits.
// When CloseForLeecher or CloseForNonMember is "1", the API is closed for those user types.
type ServerInfo struct {
	APIAccess             string `json:"apiacces"`              // Number of API accesses in the current day (GMT+1)
	CloseForLeecher       string `json:"closeforleecher"`       // API closed for non-participating members (0: open / 1: closed)
	CloseForNonMember     string `json:"closefornomember"`      // API closed for anonymous users (0: open / 1: closed)
	CPU1                  string `json:"cpu1"`                  // CPU usage % of server 1 (average of last 5 minutes)
	CPU2                  string `json:"cpu2"`                  // CPU usage % of server 2 (average of last 5 minutes)
	CPU3                  string `json:"cpu3"`                  // CPU usage % of server 3 (average of last 5 minutes)
	CPU4                  string `json:"cpu4"`                  // CPU usage % of server 4 (average of last 5 minutes)
	MaxThreadForMember    string `json:"maxthreadformember"`    // Maximum number of threads opened for members simultaneously by the API
	MaxThreadForNonMember string `json:"maxthreadfornonmember"` // Maximum number of threads opened for anonymous users simultaneously by the API
	NbScrapeurs           string `json:"nbscrapeurs"`           // Number of scrapers using the API since the last minute
	ThreadForMember       string `json:"threadformember"`       // Current number of threads opened by members simultaneously by the API
	ThreadForNonMember    string `json:"threadfornonmember"`    // Current number of threads opened by anonymous users simultaneously by the API
	ThreadsMin            string `json:"threadsmin"`            // Number of API accesses since the last minute
}

// UserInfo contains user quota and contribution information.
// Contribution level: 2 = 1 Additional Thread, 3 and + = 5 Additional Threads.
// Quota management by software is mandatory to avoid saturating servers.
type UserInfo struct {
	ID                  string `json:"id"`                  // User's username on ScreenScraper
	NumID               string `json:"numid"`               // Numeric identifier of the user on ScreenScraper
	Level               string `json:"niveau"`              // User's level on ScreenScraper
	Contribution        string `json:"contribution"`        // Financial contribution level (2 = 1 Additional Thread / 3 and + = 5 Additional Threads)
	UploadSystem        string `json:"uploadsysteme"`       // Counter of valid contributions (system media) proposed by the user
	UploadInfos         string `json:"uploadinfos"`         // Counter of valid contributions (text info) proposed by the user
	ROMAsso             string `json:"romasso"`             // Counter of valid contributions (ROM association) proposed by the user
	UploadMedia         string `json:"uploadmedia"`         // Counter of valid contributions (game media) proposed by the user
	PropositionOK       string `json:"propositionok"`       // Number of user proposals validated by a moderator
	PropositionKO       string `json:"propositionko"`       // Number of user proposals rejected by a moderator
	QuotaRefu           string `json:"quotarefu"`           // Percentage of proposal rejection by the user
	MaxThreads          string `json:"maxthreads"`          // Number of threads allowed for the user
	MaxDownloadSpeed    string `json:"maxdownloadspeed"`    // Download speed (in KB/s) allowed for the user
	RequestsToday       string `json:"requeststoday"`       // Total number of API calls during the current day
	RequestsKOToday     string `json:"requestskotoday"`     // Number of API calls with negative return (ROM/game not found) during the current day
	MaxRequestsPerMin   string `json:"maxrequestspermin"`   // Maximum number of API calls allowed per minute for the user (see FAQ)
	MaxRequestsPerDay   string `json:"maxrequestsperday"`   // Maximum number of API calls allowed per day for the user (see FAQ)
	MaxRequestsKOPerDay string `json:"maxrequestskoperday"` // Maximum number of API calls with negative return (ROM/game not found) allowed per day for the user (see FAQ)
	Visites             string `json:"visites"`             // Number of user visits to ScreenScraper
	LastVisitDate       string `json:"datedernierevisite"`  // Date of the user's last visit to ScreenScraper (format: yyyy-mm-dd hh:mm:ss)
	FavoriteRegion      string `json:"favregion"`           // Favorite region of user visits to ScreenScraper (france, europe, usa, japan)
}

// Media is a common media descriptor used across multiple endpoints
type Media struct {
	CRC     string `json:"crc,omitempty"`
	MD5     string `json:"md5,omitempty"`
	SHA1    string `json:"sha1,omitempty"`
	Format  string `json:"format,omitempty"`
	Parent  string `json:"parent,omitempty"`
	Region  string `json:"region,omitempty"`
	Size    string `json:"size,omitempty"`
	Support string `json:"support,omitempty"`
	Type    string `json:"type"`
	URL     string `json:"url,omitempty"`
}

// LocalizedName represents a name in a specific language
type LocalizedName struct {
	Language string `json:"langue"`
	Text     string `json:"text"`
}

// NameEntry represents a name entry with region and text
type NameEntry struct {
	Region string `json:"region"`
	Text   string `json:"text"`
}
