# PermCalc

This is a sub-package of DiscordConsole you can use in your own applications!  
It's a simple Discord permission calculator.

It's simple to use:
```Go
perm, err := permcalc.Show()
```

`perm` is the permission as an integer. For example 8 for admin.  
`err` is any error encountered during the process.

You can also pre-set permissions...
```Go
pm := permcalc.PermCalc{
	Perm: 8,
}
err := pm.Show()
perm := pm.Perm // Updated by Show()
```

... or even make it read-only!
```Go
pm := permcalc.PermCalc{
	ReadOnly: true,
}
err := pm.Show()
```
