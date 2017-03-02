package main;

import (
	"github.com/chzyer/readline"
)

func setCompleter(rl *readline.Instance){
	rl.Config.AutoComplete = readline.NewPrefixCompleter(
		readline.PcItem("guild", readline.PcItemDynamic(searchMap(&cacheGuilds))),
		readline.PcItem("channel", readline.PcItemDynamic(searchMap(&cacheChannels))),

		readline.PcItem("edit", readline.PcItemDynamic(singleValue(&lastUsedMsg))),
		readline.PcItem("del", readline.PcItemDynamic(singleValue(&lastUsedMsg))),
		readline.PcItem("quote", readline.PcItemDynamic(singleValue(&lastUsedMsg))),
		readline.PcItem("reactadd", readline.PcItemDynamic(singleValue(&lastUsedMsg))),
		readline.PcItem("reactdel", readline.PcItemDynamic(singleValue(&lastUsedMsg))),

		readline.PcItem("roleedit", readline.PcItemDynamic(singleValue(&lastUsedRole))),
		readline.PcItem("roledelete", readline.PcItemDynamic(singleValue(&lastUsedRole))),

		readline.PcItem("roleadd", readline.PcItem(ID,
			readline.PcItemDynamic(singleValue(&lastUsedRole)),
		)),
		readline.PcItem("roledel", readline.PcItem(ID,
			readline.PcItemDynamic(singleValue(&lastUsedRole)),
		)),
	);
}
func searchMap(theMap *map[string]string) func(string) []string{
	return func(line string) []string{
		items := make([]string, 0);
		for key := range *theMap{
			items = append(items, key);
		}
		return items;
	};
}
func singleValue(val *string) func(string) []string{
	return func(line string) []string{
		return []string{ *val };
	};
}
