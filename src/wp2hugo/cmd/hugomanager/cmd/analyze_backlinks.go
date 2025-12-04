package cmd

import (
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/backlinkanalyzer"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const _downloadInstructions = `
To download the backlinks from Bing Webmaster Tools, follow these steps:
1. Go to Bing Webmaster Tools.
2. Navigate to the Backlinks section.
3. Click on the "Pages" tab.
4. Download all backlinks data.

Alternatively, go to https://www.bing.com/webmasters/backlinks?activeTab=pages and click
"Download All" to get the backlinks data for your site.
`

func init() {
	var bingBacklinksFilepath string // See _downloadInstructions
	var colorLogOutput bool
	var numDomains *int
	cmd := &cobra.Command{
		Use:   "analyze-backlinks",
		Short: "Analyzes backlinks and shows good quality backlinks",
		Long:  "Analyzes backlinks and shows good quality backlinks.\n" + _downloadInstructions,
		Run: func(cmd *cobra.Command, args []string) {
			logger.ConfigureLogging(colorLogOutput)
			analyzeBacklinks(bingBacklinksFilepath, *numDomains)
		},
	}

	cmd.Flags().StringVarP(&bingBacklinksFilepath, "file", "", "", "Path to the Bing backlinks file"+_downloadInstructions)
	cmd.PersistentFlags().BoolVarP(&colorLogOutput, "color-log-output", "", true,
		"enable colored log output, set false to structured JSON log")
	numDomains = cmd.PersistentFlags().IntP("num-domains", "n", 10, "number of top domains to show in the output")
	rootCmd.AddCommand(cmd)
}

func analyzeBacklinks(bingBacklinksFilepath string, numDomainsToOutput int) {
	if bingBacklinksFilepath == "" {
		log.Fatal().Msg("Bing backlinks file path is required")
	}

	log.Info().
		Str("bingBacklinksFilepath", bingBacklinksFilepath).
		Msg("Analyzing backlinks...")
	analyzer, err := backlinkanalyzer.NewBingBacklinkAnalyzer(bingBacklinksFilepath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create BingBacklinkAnalyzer")
	}

	result, err := analyzer.AnalyzeBacklinks()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to analyze backlinks")
	}

	log.Debug().
		Int("numBacklinks", len(result)).
		Int("numReferringPages", result.NumReferringPages()).
		Msg("Backlinks analysis completed")

	domainFrequency, err := result.HostAndFrequency(numDomainsToOutput)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get host and frequency from backlinks")
	}

	for _, domainAndFrequency := range domainFrequency {
		log.Info().
			Str("domain", domainAndFrequency.Key).
			Int("frequency", domainAndFrequency.Value).
			Msg("Domain frequency")
	}

	// for webURL, backlinks := range result {
	//	for backlink, _ := range backlinks {
	//		fmt.Printf("%s, %s\n", webURL, backlink)
	//	}
	//}
}
