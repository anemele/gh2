package download

import (
	"fmt"

	. "gh2/pkg/rest"

	"charm.land/huh/v2"
)

func SelectReleases(
	repo Repo,
	releases []Release,
) (sels []Asset, err error) {
	var options1 []huh.Option[*Release]
	for _, release := range releases {
		title := release.Title()
		options1 = append(options1,
			huh.NewOption(title, &release))
	}

	var selectedRelease *Release
	err = huh.NewSelect[*Release]().
		Filtering(true).
		Options(options1...).
		Value(&selectedRelease).
		Height(6).
		Run()
	if err != nil {
		return
	}

	var options2 []huh.Option[Asset]
	for _, asset := range selectedRelease.Assets {
		title := asset.Title()
		options2 = append(options2,
			huh.NewOption(title, asset))
	}

	err = huh.NewMultiSelect[Asset]().
		Title("select assets").
		Filterable(true).
		Options(options2...).
		Value(&sels).
		Height(6).
		Run()
	if err != nil {
		sels = nil
		return
	}

	if len(sels) == 0 {
		sels = nil
		err = fmt.Errorf("nothing selected")
		return
	}

	return sels, nil
}

func SelectCache(repos []string) (
	sels []string, err error,
) {
	var options []huh.Option[string]

	for _, repo := range repos {
		options = append(options, huh.NewOption(repo, repo))
	}

	err = huh.NewMultiSelect[string]().
		Title("select repos").
		Filterable(true).
		Options(options...).
		Value(&sels).
		Height(6).
		Run()
	if err != nil {
		sels = nil
		return
	}
	if len(sels) == 0 {
		err = fmt.Errorf("nothing selected")
		return
	}
	return sels, nil
}
