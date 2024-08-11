package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Border(lipgloss.RoundedBorder())

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

type item struct {
	title  string
	artist string
}

type trackDetail struct {
	directory string
	fileExt   int
}

type tickMsg time.Time

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.artist }
func (i item) FilterValue() string { return i.title }

type listKeyMap struct {
	toggleTitleBar  key.Binding
	toggleStatusBar key.Binding
	toggleHelpMenu  key.Binding
	togglePlay      key.Binding
	toggleMute      key.Binding
	volumeUp        key.Binding
	volumeDown      key.Binding
	toggleLoop      key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		toggleTitleBar: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "toggle title"),
		),
		toggleStatusBar: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "toggle status"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
		togglePlay: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "toggle playback"), // change this from "u" to Space, figure out how binding space works
		),
		toggleMute: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "mute playback"),
		),
		volumeUp: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "volume up"),
		),
		volumeDown: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "volume down"),
		),
		toggleLoop: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "toggle looping"),
		),
	}
}

type model struct {
	list         list.Model
	keys         *listKeyMap
	delegateKeys *delegateKeyMap
	progress     progress.Model
}

func newModel() model {
	var (
		delegateKeys = newDelegateKeyMap()
		listKeys     = newListKeyMap()
		progress     = progress.New(progress.WithDefaultGradient())
	)

	items := make([]list.Item, len(testList))
	i := 0
	for k := range testList {
		items[i] = item{
			title:  k.title,
			artist: k.artist,
		}
		i++
	}

	// Setup list
	delegate := newItemDelegate(delegateKeys)
	songList := list.New(items, delegate, 0, 0)
	songList.Title = "Songs"
	songList.Styles.Title = titleStyle
	songList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleTitleBar,
			listKeys.toggleStatusBar,
			listKeys.toggleHelpMenu,
			listKeys.togglePlay,
			listKeys.toggleMute,
		}
	}

	return model{
		list:         songList,
		keys:         listKeys,
		delegateKeys: delegateKeys,
		progress:     progress,
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tickMsg:
		if m.progress.Percent() == 1.0 {
			return m, nil
		}

		cmd := m.progress.IncrPercent(0.25)
		return m, tea.Batch(tickCmd(), cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.toggleTitleBar):
			v := !m.list.ShowTitle()
			m.list.SetShowTitle(v)
			m.list.SetShowFilter(v)
			m.list.SetFilteringEnabled(v)
			return m, nil

		case key.Matches(msg, m.keys.toggleStatusBar):
			m.list.SetShowStatusBar(!m.list.ShowStatusBar())
			return m, nil

		case key.Matches(msg, m.keys.toggleHelpMenu):
			m.list.SetShowHelp(!m.list.ShowHelp())
			return m, nil

		case key.Matches(msg, m.keys.togglePlay):
			togglePlay()
			return m, nil

		case key.Matches(msg, m.keys.toggleMute):
			toggleMute()
			return m, nil
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return appStyle.Render(m.list.View())
}

func listDir(root string) (map[item]*trackDetail, error) {
	var files = make(map[item]*trackDetail)
	err := filepath.Walk(root, func(location string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			switch filepath.Ext(location) {
			case ".flac":
				if strings.LastIndex(location, "/") >= 0 {
					dir, file := path.Split(location)
					tempItem := item{file, "1"}
					files[tempItem] = &trackDetail{dir, 1}
				}
			case ".wav":
				if strings.LastIndex(location, "/") >= 0 {
					dir, file := path.Split(location)
					tempItem := item{file, "2"}
					files[tempItem] = &trackDetail{dir, 2}
				}
			case ".mp3":
				if strings.LastIndex(location, "/") >= 0 {
					dir, file := path.Split(location)
					tempItem := item{file, "3"} // there has to be a better way for this
					files[tempItem] = &trackDetail{dir, 3}
				}
			default:
				return nil
			}
		}
		return nil
	})
	return files, err
}

var testList = make(map[item]*trackDetail)

func main() {
	var directory string
	var err error
	if len(os.Args) < 2 {
		directory, err = os.UserHomeDir()
	} else {
		directory = os.Args[1]
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	testList, err = listDir(directory)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
