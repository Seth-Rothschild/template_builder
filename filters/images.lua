local M = {}

function M.source_dir()
  local input_file = PANDOC_STATE.input_files[1]
  if input_file == nil then
    return nil
  end
  return input_file:match("^(.*)/[^/]+$")
end

function M.disk_path(src)
  if src:sub(1, 1) == "/" then
    return src
  end
  local source_dir = M.source_dir()
  if source_dir == nil then
    return src
  end
  return source_dir .. "/" .. src
end

function M.image_exists(src)
  local file = io.open(src, "rb")
  if file == nil then
    return false
  end
  file:close()
  return true
end

function M.find_image(fig)
  for _, block in ipairs(fig.content) do
    for _, inline in ipairs(block.content) do
      if inline.t == "Image" then
        return inline
      end
    end
  end
  return nil
end

function M.render_caption(fig)
  local doc = pandoc.Pandoc(fig.caption.long)
  local latex = pandoc.write(doc, "latex")
  return latex:gsub("%s+$", "")
end

function M.build_figure(fig)
  local image = M.find_image(fig)
  local caption = M.render_caption(fig)

  local found = false
  if image ~= nil then
    found = M.image_exists(M.disk_path(image.src))
  end

  local lines = {}
  table.insert(lines, "\\begin{figure}[H]")
  table.insert(lines, "\\centering")
  if found then
    table.insert(lines, "\\includegraphics[width=\\linewidth,height=\\textheight,keepaspectratio]{" .. image.src .. "}")
    table.insert(lines, "\\caption{" .. caption .. "}")
  else
    table.insert(lines, "\\refstepcounter{figure}Figure \\thefigure{} (missing image): " .. caption)
  end
  table.insert(lines, "\\end{figure}")

  return table.concat(lines, "\n")
end

function M.Figure(fig)
  return pandoc.RawBlock("latex", M.build_figure(fig))
end

return { M }
