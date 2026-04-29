package render

func (s *Shell) moveSelection(delta int) bool {
	section := s.listSection()
	if !section.Scrollable() || delta == 0 {
		return false
	}

	data := s.listDataForSection(section)
	if data == nil || len(data.Items) == 0 {
		return false
	}

	current := s.currentSelectionIndex(section)
	next := clampInt(current+delta, 0, len(data.Items)-1)
	if next == current {
		return false
	}

	s.setSelection(section, next)
	return true
}

func (s *Shell) moveSelectionTo(index int) bool {
	section := s.listSection()
	if !section.Scrollable() {
		return false
	}

	data := s.listDataForSection(section)
	if data == nil || len(data.Items) == 0 {
		return false
	}

	if index > len(data.Items)-1 {
		index = len(data.Items) - 1
	}
	index = clampInt(index, 0, len(data.Items)-1)

	if s.currentSelectionIndex(section) == index {
		return false
	}

	s.setSelection(section, index)
	return true
}

func (s *Shell) pageSize() int {
	size := s.viewHeights[s.listSection()]
	if size <= 1 {
		return 8
	}

	return size - 1
}

func (s *Shell) currentSelectedItem(section Section) *ListItemData {
	data := s.listDataForSection(section)
	if data == nil || len(data.Items) == 0 {
		return nil
	}

	index := s.currentSelectionIndex(section)
	if index < 0 || index >= len(data.Items) {
		return nil
	}

	return &data.Items[index]
}

func (s *Shell) currentSelectionIndex(section Section) int {
	if !section.Scrollable() {
		return -1
	}

	s.reconcileSelection(section)

	data := s.listDataForSection(section)
	if data == nil || len(data.Items) == 0 {
		return -1
	}

	index := s.selectedIndexes[section]
	if index < 0 || index >= len(data.Items) {
		index = 0
		s.setSelection(section, index)
	}

	return index
}

func (s *Shell) setSelection(section Section, index int) {
	data := s.listDataForSection(section)
	if data == nil || len(data.Items) == 0 {
		delete(s.selectedIndexes, section)
		delete(s.selectedKeys, section)
		return
	}

	index = clampInt(index, 0, len(data.Items)-1)
	s.selectedIndexes[section] = index
	s.selectedKeys[section] = data.Items[index].Key
}

// scrollToSelected adjusts s.scrolls[section] so the currently selected
// item is visible. Used after programmatic navigation (search confirm, jump
// to entry) where the scroll position was not updated alongside the selection.
func (s *Shell) scrollToSelected(section Section) {
	index := s.selectedIndexes[section]
	pageH := s.viewHeights[section]
	if pageH <= 0 {
		pageH = 8 // conservative fallback before first render
	}

	scroll := s.scrolls[section]
	if index < scroll {
		scroll = index
	} else if index >= scroll+pageH {
		scroll = index - pageH + 1
	}
	if scroll < 0 {
		scroll = 0
	}
	s.scrolls[section] = scroll
}

func (s *Shell) reconcileSelections() {
	for _, section := range orderedSections {
		if !section.Scrollable() {
			continue
		}
		s.reconcileSelection(section)
	}
	// Settings tabs are scrollable but not in orderedSections.
	for _, tab := range settingsTabs {
		s.reconcileSelection(tab)
	}
}

func (s *Shell) reconcileSelection(section Section) {
	data := s.listDataForSection(section)
	if data == nil || len(data.Items) == 0 {
		delete(s.selectedIndexes, section)
		delete(s.selectedKeys, section)
		s.scrolls[section] = 0
		return
	}

	if key := s.selectedKeys[section]; key != "" {
		if index := listItemIndexByKey(data.Items, key); index >= 0 {
			s.selectedIndexes[section] = index
			return
		}
	}

	index := s.selectedIndexes[section]
	if index < 0 || index >= len(data.Items) {
		index = 0
	}

	s.setSelection(section, index)
}

func listItemIndexByKey(items []ListItemData, key string) int {
	for index, item := range items {
		if item.Key == key {
			return index
		}
	}

	return -1
}
