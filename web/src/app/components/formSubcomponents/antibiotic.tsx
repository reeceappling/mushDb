// non-client

import {Data, GroupProps} from "@/app/components/formSubcomponents/shared";
import {SelectorResetsOnSelectFor} from "@/app/components/selector";
import {useQuery} from "@tanstack/react-query";
import {getOptionsResponse} from "@/app/components/formSubcomponents/server";

export type Antibiotic = "Doxycycline" | "Cefazolin" | "Amoxicillin";
export const AntibioticsList: Antibiotic[] = ["Doxycycline", "Cefazolin", "Amoxicillin"]

export function IsValidAntibiotic(input: any): boolean {
    return (
        typeof input === 'string' && AntibioticsList.includes(input as Antibiotic) // TODO: fixme
    )
}

export function AntibioticSelector( // TODO: can only disable/enable, delete will happen on update
    {initial, onSelect, blacklist}: {
        initial?: Antibiotic,
        onSelect?: (ab?: Antibiotic) => void,
        blacklist?: string[],
    }) {

    // TODO: I think that something weird is going on with the

    const {isPending, error, data} = useQuery({ // TODO: SO SOMETHING SIMILAR FOR nutrients, sugars, additives, liquid, grain, transferReason
        queryKey: ['antibioticOptions'],
        queryFn: () => getOptionsResponse("antibiotics")
    })
    if (isPending || error !== null) {
        return <div>{isPending ? "ANTIBIOTIC SELECTOR LOADING" : "ANTIBIOTIC SELECTOR ERROR: " + error.message}</div>
    }
    const filteredOptions = data.filter((val, idx) => {
        return !(blacklist && blacklist.includes(val))
    })
    return <SelectorResetsOnSelectFor options={["", ...filteredOptions]} updateParent={(s) => {
        if (s === "") {
            onSelect && onSelect()
        }
        onSelect && onSelect(s as Antibiotic)
    }}/>
}

export function AntibioticsDisplay( // TODO: can only disable/enable, delete will happen on update
    {antibiotics}: {
        antibiotics?: Antibiotic[],
    }) {
    if (!antibiotics || antibiotics.length === 0) {
        return <div>
            <div>{"Antibiotics: None"}</div>
        </div>
    }
    return <div>
        <div>{"Antibiotics:"}</div>
        {antibiotics.map((ab) => {
            return <div key={ab}>{ab}</div> // TODO: LINK TO INFO/DOSAGE?
        })}
    </div>
}

export function AntibioticEntriesGroupForNew(
    {currentEntries, updateParent,}: {
        currentEntries: Antibiotic[],
        updateParent: (l: Antibiotic[]) => void
    }) {
    const addEntryByName = (item?: Antibiotic) => {
        item && updateParent([...(currentEntries || []), item])
    }
    return <div>{/* TODO: CLASS STYLINGS!!!! */}
        {currentEntries.map((input, index) => {
            return <div key={index} className={"inlineChildren mb-1"}> {/* TODO: CLASS STYLINGS!!!! */}
                <div>{input}</div>
                <button className={"removeButton ml-2"} onClick={()=>{
                    updateParent([...(currentEntries || [])].filter((existing) => existing !== input))
                }}>{"Remove"}</button>
            </div>
        })}
        <AntibioticSelector onSelect={addEntryByName} blacklist={currentEntries.map((v) => {
            return v
        })}/>
    </div>
}

export function AntibioticEntriesGroup({
                                           initialEntries,

                                           preexisting,
                                           readonly,
                                           updateParent,
                                           blacklist
                                       }: GroupProps<Antibiotic>) {
    const removeFields = (toRemove: string) => {
        updateParent([...(initialEntries || [])].filter((c, i) => c.data !== toRemove))
    }
    const disableField = (name: string) => { // TODO: unsure if this works properly
        const data = [...(initialEntries || [])].map((v, i) => {
            const val = v
            if (val.data === name) {
                val.disabled = !val.disabled
            }
            return val
        });
        updateParent(data)
    }

    const groupClasses = () => {
        let out = ""
        if (preexisting) {
            out = "exists"
        } else {
            out = "new"
        }
        if (readonly) {
            out += " readonly"
        } else {
            out += " editable"
        }
        return out
    }
    const entryClasses = (entry: Data<Antibiotic>) => {
        let out = "antibioticEntry"
        if (entry.disabled) {
            out += " disabled"
        } else {
            out += " enabled"
        }
        return out
    }
    const addEntryByName = (item?: Antibiotic) => {
        if (item === undefined) {
            return
        }
        let data = [...(initialEntries || [])];
        const i = data.findIndex((dataItem) => {
            return dataItem.data === item
        })
        if (i !== -1) {
            if (!data[i].disabled) {
                console.log("field already enabled")
            } else {
                console.log("field reenabled")
                data[i].disabled = false
            }

        } else {
            console.log("field should be added")
            data = [...data, {data: item, disabled: false}]
        }

        updateParent(data)
    }
    // TODO: match this up with AdditiveEntriesGroup
    return <div className={groupClasses()}>{/* TODO: CLASS STYLINGS!!!! */}
        {readonly || <div>
            <AntibioticSelector onSelect={addEntryByName} blacklist={initialEntries?.map((v) => {
                return v.data
            })}/>{/* TODO: RESET THE SELECTOR ON SELECT*/}
        </div>}
        {initialEntries?.map((input, index) => {
            return <div className={"editable "+entryClasses(input)} key={index}> {/* TODO: CLASS STYLINGS!!!! */}
                {input.disabled ? "disabled" : null /* TODO: remove! */}
                {initialEntries[index].data}
                {(!readonly) && <button className={input.disabled?"basicButton":"removeButton"} onClick={() => {
                    (preexisting ? disableField(input.data) : removeFields(input.data))
                }}>
                    {preexisting ? (input.disabled ? "enable" : "disable") : "remove"}
                </button>}
            </div>
        })}
    </div>
}