package core

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
)

func SurveyReleases(repo *Repo, releases []Release) ([]Asset, error) {
	var releaseTitles []string
	var releaseTitleMap = make(map[string]Release)
	for _, release := range releases {
		title := release.Title()
		releaseTitles = append(releaseTitles, title)
		releaseTitleMap[title] = release
	}

	// ask select releases
	var q1 = &survey.Select{
		Message: fmt.Sprintf("Select releases of %s:", repo.String()),
		Options: releaseTitles,
	}
	var selectedReleaseName string
	err := survey.AskOne(q1, &selectedReleaseName)
	if err != nil {
		return nil, err
	}

	if selectedReleaseName == "" {
		return nil, fmt.Errorf("nothing selected")
	}
	selectedRelease := releaseTitleMap[selectedReleaseName]

	var assetTitles []string
	var assetTitleMap = make(map[string]Asset)
	for _, asset := range selectedRelease.Assets {
		title := asset.Title()
		assetTitles = append(assetTitles, title)
		assetTitleMap[title] = asset
	}

	// ask select assets
	var q2 = &survey.MultiSelect{
		Message: "Select assets:",
		Options: assetTitles,
	}
	var selectedAssetNames []string
	err = survey.AskOne(q2, &selectedAssetNames)
	if err != nil {
		return nil, err
	}
	if len(selectedAssetNames) == 0 {
		return nil, fmt.Errorf("nothing selected")
	}
	var selectedAssets []Asset
	for _, assetName := range selectedAssetNames {
		asset := assetTitleMap[assetName]
		selectedAssets = append(selectedAssets, asset)
	}

	return selectedAssets, nil
}

func SurveyCache(repos []string) ([]string, error) {
	// ask select repos
	var q1 = &survey.MultiSelect{
		Message: "Select repos:",
		Options: repos,
	}
	var selected []string
	err := survey.AskOne(q1, &selected)
	if err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("nothing selected")
	}
	return selected, nil
}
