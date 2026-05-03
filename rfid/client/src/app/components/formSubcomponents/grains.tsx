// TODO: THESE ARE ALL NON-CLIENT!

import {useQuery} from "@tanstack/react-query";
import {SelectorFor} from "@/app/components/selector";
import {getOptionsResponse} from "@/app/components/formSubcomponents/server";

export const GrainsList: string[] = ["Oats", "Popcorn", "Wheat", "Rye", "Millett"]



export function GrainsTypeSelector( // TODO: USE THIS!!!!!
    {initial, onSelect,blacklist}:{
        initial?: string,
        onSelect?: (ab?: string)=>void
        blacklist?: string[],
    }){
    const { isPending, error, data } = useQuery({
        queryKey: ['grainsOptions'],
        queryFn: () => getOptionsResponse("grains")
    })
    if(isPending || error !== null){
        return <div>{isPending ? "GRAIN SELECTOR LOADING" : "GRAIN SELECTOR ERROR: "+error.message}</div>
    }
    const filteredOptions = data.filter((val, idx)=>{
        return !(blacklist && blacklist.includes(val))
    })
    return <SelectorFor disabled={onSelect===undefined} options={["", ...filteredOptions]} initial={initial || ""} updateParent={(s)=>{
        if(s===""){
            onSelect && onSelect()
        }
        onSelect && onSelect(s)}
    } />
}

export function GrainsSelector({current, onChange}:{current: Grain[], onChange: (gs: Grain[])=>void}){
    const grainTypeAmountSelectors = ()=>{
        return <div>
            {current.map((g, idx)=>{
                return <div>
                    <GrainsTypeSelector initial={g.grain} onSelect={(gr)=>{
                        if(gr){
                            let newGs = [...current]
                            newGs[idx].grain = gr
                            onChange(newGs)
                        }
                    }}/>
                    {/* TODO INPUT NUMBER */}
                    {/* TODO REMOVER */}
                </div>
            })}
        </div>
    }
    const addGrain = (g?: string) => {
        if(g){
            onChange([...current, {grain: g, percentage: 0}])
        }
    }
    return <div>
        {grainTypeAmountSelectors()}
        <GrainsTypeSelector onSelect={addGrain}/>
        {"FIXME!!!!!"}
    </div>
}

export interface Grain {
    grain: string,
    percentage: number,
}

export function IsValidGrain(input: any): boolean {
    return (
        typeof input === 'object' &&
        'grain' in input && typeof input.grain === 'string' &&
        'percentage' in input && typeof input.percentage === 'number'
    )
}