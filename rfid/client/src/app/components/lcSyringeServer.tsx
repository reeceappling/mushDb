import {Note} from "@/app/components/formSubcomponents/notes";
import {ACL} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {ChannelTextNewAgarBatch} from "@/app/components/agarBatchServer";
import {LcSelector} from "@/app/components/lcClient";
import {LcData} from "@/app/components/lcServer";
import {LcSyringeSelector} from "@/app/components/lcSyringeClient";

export function TestLcSyringeOk(){
    let ExampleNotes;
    const a: LcSyringe = {
        _id: "(LC ID HERE)",
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        confirmedClean: undefined,
        knownFruitable: true,
        genSpore: 7,
        genFruitOrSpore: 3,
        transfersOut: ["(TRANSFER 1)","(TRANSFER 2)"],
        parent: "(PARENT ID)",
        disposed: Date.now()+40000,
        notes: ExampleNotes,
        lastUpdated: 789,
    }
    return a
}
export interface LcSyringe {
    _id: string
    parent?: string
    creationDate: number
    species: string
    subspecies?: string
    sale?: string
    genSpore?:  number
    genFruitOrSpore?: number
    confirmedClean?: boolean
    knownFruitable?: boolean
    transfersOut?: string[]
    disposed?: number
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}

export function LcSyringeSelectorCloseable(sp: SelectorProps<LcSyringe>) { // TODO: use
    const doSel = (val?: LcSyringe):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<LcSyringe> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        msgTxt: ChannelTextNewAgarBatch, // TODO: ???
        closeTxt: "Close LC Syringe List",
        //createTxt: "Create Bag",// TODO: ???
        lowercase: "liquid culture syringe",
        //creatorInPage: sp.creatorInPage,// TODO: ???
        //createEndpt: "bag",// TODO: ???
        getId: (v: LcSyringe) => v._id,
        createSelector:(selHdl: (onSelect: LcSyringe) => void)=>{
            return <LcSyringeSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: ok?
        // createCreator:(selHdl: (onSelect: FruitingChamberData) => void)=>{
        //     return <NewFruitingChamberForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}