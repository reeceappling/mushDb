// serverside. No state

import {useQuery} from "@tanstack/react-query";
import {getOptionsResponse} from "@/app/components/formSubcomponents/server";
import {SelectorResetsOnSelectFor} from "@/app/components/selector";
import * as React from "react";
import {DowelEntryForNew, RemoveButton} from "@/app/components/formSubcomponents/commonClient";
import {DowelType} from "@/app/components/plugsServer";

export function WoodTypeSelectorForNew(
    {onSelect, blacklist}: {
        onSelect?: (ab?: string) => void,
        blacklist?: string[],
    }) {
    const {isPending, error, data} = useQuery({
        queryKey: ['woodsOptions'],
        queryFn: () => getOptionsResponse("woods")
    })
    if (isPending || error !== null) {
        return <div>{isPending ? "Loading wood selector " : "Wood selector error: " + error.message}</div>
    }
    const filteredOptions = data.filter((val, idx) => {
        return !(blacklist && blacklist.includes(val))
    })
    return <SelectorResetsOnSelectFor options={["", ...filteredOptions]} updateParent={(s) => {
        if (s === "") {
            onSelect && onSelect()
        }
        onSelect && onSelect(s as string)
    }
    }/>
}

export function WoodEntriesGroupForNew({currentEntries, updateParent}: {
    currentEntries: DowelType[],
    updateParent: (l: DowelType[]) => void
}) {
    const handleSelect = (v: DowelType) => {
        const data = [...(currentEntries || []), v];
        updateParent(data)
    }
    return <div>
        {currentEntries.length !== 0 && <div className={"inputGrid inputGrid4 gap-8"}>
            {currentEntries.map((n, i) => {
                const keepOtherWoods = (existing: DowelType)=>existing.wood !== n.wood
                return <div key={n.wood} className={"contentsOnly"}>
                    <DowelEntryForNew initial={{wood:n.wood,size:0.25,units:"in"}} updateParent={(updated: DowelType) => {
                        updateParent([...(currentEntries || [])].map((existing: DowelType)=>{
                            return existing.wood !== n.wood ? existing : updated
                        }))
                    }}/>
                    <RemoveButton txt={"Remove"} click={() => {
                        updateParent([...(currentEntries || [])].filter(keepOtherWoods))
                    }}/>
                </div>
            })}
        </div>}
        <WoodTypeSelectorForNew onSelect={(val) => {
            val && handleSelect({wood: val, size: 1.0/4.0, units: "in"}) // TODO: default ok?
        }} blacklist={currentEntries.map((v) => v.wood)}/>
    </div>
}