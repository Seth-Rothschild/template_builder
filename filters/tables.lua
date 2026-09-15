local M = {}

function M.render_cell(cell)
  local doc = pandoc.Pandoc(cell.contents)
  local latex = pandoc.write(doc, "latex")
  return latex:gsub("%s+$", "")
end

function M.render_row(row)
  local cells = {}
  for _, cell in ipairs(row.cells) do
    table.insert(cells, (M.render_cell(cell)))
  end
  return table.concat(cells, " & ") .. " \\\\"
end

function M.build_table(tbl)
  local column_count = #tbl.colspecs
  local column_spec = string.rep("X[l]", column_count)

  local rows = {}
  for _, row in ipairs(tbl.head.rows) do
    table.insert(rows, M.render_row(row))
  end
  for _, body in ipairs(tbl.bodies) do
    for _, row in ipairs(body.body) do
      table.insert(rows, M.render_row(row))
    end
  end

  local lines = {}
  table.insert(lines, "\\begin{longtblr}{")
  table.insert(lines, "  colspec = {" .. column_spec .. "},")
  table.insert(lines, "  row{1} = {bg=accent, fg=white, font=\\bfseries},")
  table.insert(lines, "  row{even} = {bg=gray!10},")
  table.insert(lines, "  hline{1,2,Z} = {solid, 0.8pt, accent},")
  table.insert(lines, "}")
  for _, row in ipairs(rows) do
    table.insert(lines, row)
  end
  table.insert(lines, "\\end{longtblr}")

  return table.concat(lines, "\n")
end

function M.Table(tbl)
  return pandoc.RawBlock("latex", M.build_table(tbl))
end

return { M }
