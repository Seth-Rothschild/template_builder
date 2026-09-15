local tables = dofile("filters/tables.lua")

local function assert_equal(expected, actual, message)
  if expected ~= actual then
    error(string.format("%s: expected %q, got %q", message, expected, actual))
  end
end

local cell = pandoc.Cell({ pandoc.Plain({ pandoc.Str("X") }) }, pandoc.AlignDefault, 1, 1)
assert_equal("X", tables.render_cell(cell), "render_cell")

local row = pandoc.Row({
  pandoc.Cell({ pandoc.Plain({ pandoc.Str("A") }) }, pandoc.AlignDefault, 1, 1),
  pandoc.Cell({ pandoc.Plain({ pandoc.Str("B") }) }, pandoc.AlignDefault, 1, 1),
})
assert_equal("A & B \\\\", tables.render_row(row), "render_row")

local one_column_table = pandoc.Table(
  pandoc.Caption(nil),
  { { pandoc.AlignDefault, pandoc.ColWidthDefault } },
  pandoc.TableHead({
    pandoc.Row({
      pandoc.Cell({ pandoc.Plain({ pandoc.Str("Header") }) }, pandoc.AlignDefault, 1, 1),
    }),
  }),
  {},
  pandoc.TableFoot({})
)
local expected_one_column_table = "\\begin{longtblr}{\n"
    .. "  colspec = {X[l]},\n"
    .. "  row{1} = {bg=accent, fg=white, font=\\bfseries},\n"
    .. "  row{even} = {bg=gray!10},\n"
    .. "  hline{1,2,Z} = {solid, 0.8pt, accent},\n"
    .. "}\n"
    .. "Header \\\\\n"
    .. "\\end{longtblr}"
assert_equal(expected_one_column_table, tables.build_table(one_column_table), "build_table")

print("all tests passed")
