import {Note} from "@/app/components/formSubcomponents/notes";
import {TestNotes} from "@/app/components/formSubcomponents/contaminations";
import {ACL} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {ChannelTextNewAgarBatch} from "@/app/components/agarBatchServer";
import {SlantSelector} from "@/app/components/slantClient";
import {SlantData} from "@/app/components/slantServer";
import {SporeSwabSelector} from "@/app/components/sporeSwabClient";


export function TestSporeSwabOk(){
    const a: SporeSwab = {
        _id: "(SUBSTR ID HERE)",
        parent: "(PARENT ID)",
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        sale: "SALE ID",
        disposed: Date.now(),
        notes: TestNotes,
        lastUpdated: 789,
    }
    return a
}

export interface SporeSwab {
    _id: string
    parent?: string // Only empty if purchased and not printed yourself
    parentType?:string
    creationDate: number
    species: string
    subspecies?: string
    sale?: string
    disposed?: number
    transfersOut?: string[]
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}

export function SporeSwabSelectorCloseable(sp: SelectorProps<SporeSwab>) { // TODO: use
    const doSel = (val?: SporeSwab):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<SporeSwab> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        msgTxt: ChannelTextNewAgarBatch, // TODO: ???
        closeTxt: "Close Spore Swab List",
        //createTxt: "Create Bag",// TODO: ???
        lowercase: "spore swab",
        //creatorInPage: sp.creatorInPage,// TODO: ???
        //createEndpt: "bag",// TODO: ???
        getId: (v: SporeSwab) => v._id,
        createSelector:(selHdl: (onSelect: SporeSwab) => void)=>{
            return <SporeSwabSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: ok?
        // createCreator:(selHdl: (onSelect: FruitingChamberData) => void)=>{
        //     return <NewFruitingChamberForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}