package download

import (
	"fmt"

	. "gh2/pkg/rest"

	"charm.land/huh/v2"
)

func MultiSelect[T comparable](
	title string,
	options []huh.Option[T],
) (sels []T, err error) {
	err = huh.NewMultiSelect[T]().
		Title(title).
		Filterable(true).
		Options(options...).
		Value(&sels).
		Height(6).
		Run()
	if err != nil {
		return
	}
	if len(sels) == 0 {
		err = fmt.Errorf("nothing selected")
	}

	return
}

func SelectReleases(
	repo Repo,
	releases []Release,
) (sels []*Release, err error) {
	var options []huh.Option[*Release]
	for _, release := range releases {
		title := release.Title()
		options = append(options,
			huh.NewOption(title, &release))
	}

	return MultiSelect(repo.String(), options)
}

func SelectAssets(
	release Release,
) (sels []Asset, err error) {
	var options []huh.Option[Asset]
	for _, asset := range release.Assets {
		title := asset.Title()
		options = append(options,
			huh.NewOption(title, asset))
	}

	return MultiSelect(release.Name, options)
}

func SelectCache(
	repos []string,
) (sels []string, err error) {
	var options []huh.Option[string]
	for _, repo := range repos {
		options = append(options, huh.NewOption(repo, repo))
	}

	return MultiSelect("select repos", options)
}
