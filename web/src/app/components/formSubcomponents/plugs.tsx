// serverside. No state

import {useQuery} from "@tanstack/react-query";
import {getOptionsResponse} from "@/app/components/formSubcomponents/server";
import {SelectorResetsOnSelectFor} from "@/app/components/selector";
import * as React from "react";
import {AdditiveEntryForNew, DowelEntryForNew, RemoveButton} from "@/app/components/formSubcomponents/commonClient";
import {DowelType} from "@/app/components/plugsServer";
import {useEffect, useState} from "react";
import {Additive, AdditiveTypeSelectorForNew} from "@/app/components/formSubcomponents/additives";

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
export function WoodEntriesGroupForNew({
                                               initial,
                                               updateParent,
                                           }: {
    initial: DowelType[],
    updateParent: (l: DowelType[]) => void
}) {
    const [current, setCurrent] = useState<DowelType[]>(initial)

    useEffect(() => {
        setCurrent(initial)
    }, [initial])

    const doUpdate = (upd: DowelType[]) => {
        setCurrent(upd)
        updateParent(upd)
    }

    const handleSelectType = (v: string) => {
        const data = [...current, { wood: v, size: 0.25, units: "in" }] // TODO: MODIFY ON OTHERS!
        doUpdate(data)
    }

    return <div>
        {current.length !== 0 && <div className={"inputGrid inputGrid4 gap-8"}>
            {current.map((n, i) => {
                return <div key={n.wood} className={"contentsOnly"}>
                    <DowelEntryForNew
                        initial={n}
                        updateParent={(updated: DowelType) => {
                            doUpdate(current.map((existing, idx) => idx === i ? updated : existing))
                        }}
                    />
                    <RemoveButton
                        txt={"Remove"}
                        click={() => {
                            doUpdate(current.filter((_, idx) => idx !== i))
                        }}
                    />
                </div>
            })}
        </div>}
        <WoodTypeSelectorForNew
            onSelect={(val) => { if (val) handleSelectType(val) }}
            blacklist={current.map((v) => v.wood)}
        />
    </div>
}