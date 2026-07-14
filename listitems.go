package main

// periodItem is one row of the year or month picker - a pre-formatted
// label ("2026    $970.00 spent" / "07 July       $970.00 spent") and the
// value (year, or "MM") selectYear/selectMonth act on. Unlike a generic
// list widget's Item, there's no separate right-aligned value column
// here: populateYearList/populateMonthList build the whole line as one
// left-to-right string, matching Python's f-string formatting exactly, so
// periodItem just carries it through.
type periodItem struct {
	value string
	label string
}
