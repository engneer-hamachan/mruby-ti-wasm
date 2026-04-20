class Poge
  include IRQ
end

h = Poge.new
h.irq(1)
h.irq(1) do |x, y|
  dbtp x
  dbtp y
end
