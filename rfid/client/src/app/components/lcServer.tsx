import {Note} from "@/app/components/formSubcomponents/notes";
import {
    Contamination, ExampleContaminations, ExamplePicsWithNotesIncoming,
} from "@/app/components/formSubcomponents/contaminations";
import {
    ExamplePicWithNotesIncoming,
    PicWithNotesIncoming
} from "@/app/components/formSubcomponents/picWithNotes";
import {EntryPerms} from "@/app/components/perms";
import {ACL} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {ChannelTextNewAgarBatch} from "@/app/components/agarBatchServer";
import {JarSelector} from "@/app/components/jarClient";
import {JarData} from "@/app/components/jarServer";
import {LcSelector} from "@/app/components/lcClient";

export function TestLcOk(){
    let ExampleNotes;
    const a: LcData = {
        _id: "(LC ID HERE)",
        recipe: "(LC RECIPE ID HERE)",
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        innoc: "(Innoc transfer id!)",
        genSpore: 7,
        genFruitOrSpore: 3,
        transfersOut: ["(TRANSFER 1)","(TRANSFER 2)"],
        parentType: "lc",
        parent: "(PARENT ID)",
        pics: ExamplePicsWithNotesIncoming,
        confirmedClean: true,
        contamination: ExampleContaminations,
        knownFruitable: true,
        disposed: Date.now()+40000,
        mostRecentImage: ExamplePicWithNotesIncoming,
        notes: ExampleNotes,
        lastUpdated: 789,
        //perms: {userPerms: {ids:[{id:"userCollId",val:"userName"}],canWrite:[true]},projectPerms: {ids:["proj1","proj2"],canWrite:[true, false]}, blanketPerms: 1},
    }
    return a
}
export interface LcData {
    _id: string
    recipe: string
    pcRun?: string
    creationDate: number
    species?: string
    subspecies?: string
    innoc?: string
    genSpore?:  number
    genFruitOrSpore?: number
    transfersOut?: string[]
    parentType?: string
    parent?: string
    pics?: PicWithNotesIncoming[]
    confirmedClean?: boolean
    contamination?: Contamination[]
    knownFruitable?: boolean
    disposed?: number
    mostRecentImage?: PicWithNotesIncoming
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}

export function LcSelectorCloseable(sp: SelectorProps<LcData>) { // TODO: use
    const doSel = (val?: LcData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<LcData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        msgTxt: ChannelTextNewAgarBatch, // TODO: ???
        closeTxt: "Close LC List",
        //createTxt: "Create Bag",// TODO: ???
        lowercase: "liquid culture",
        //creatorInPage: sp.creatorInPage,// TODO: ???
        //createEndpt: "bag",// TODO: ???
        getId: (v: LcData) => v._id,
        createSelector:(selHdl: (onSelect: LcData) => void)=>{
            return <LcSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: ok?
        // createCreator:(selHdl: (onSelect: FruitingChamberData) => void)=>{
        //     return <NewFruitingChamberForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}