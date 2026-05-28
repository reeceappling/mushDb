import {ExamplePicWithNotesIncoming, PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {Note} from "@/app/components/formSubcomponents/notes";
import {EntryPerms} from "@/app/components/perms";
import {ExamplePicsWithNotesIncoming, TestNotes} from "@/app/components/formSubcomponents/contaminations";
import {ACL} from "@/app/components/accessControlServer";
import CloseableSelector from "@/app/components/selector";
import {ChannelTextNewAgarBatch} from "@/app/components/agarBatchServer";
import {FruitSelector} from "@/app/components/fruitClient";
import {FruitData} from "@/app/components/fruitServer";
import {SporePrintSelector} from "@/app/components/sporePrintClient";


export function TestSporePrintOk(){
    const a: SporePrintData = {
        _id: "(SUBSTR ID HERE)",
        parent: "(PARENT ID)",
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        color: "Black",
        density: "Average",
        pics: ExamplePicsWithNotesIncoming,
        sale: "SALE ID",
        disposed: Date.now(),
        mostRecentImage: ExamplePicWithNotesIncoming,
        notes: TestNotes,
        lastUpdated: 789,
        //perms: {userPerms: {ids:[{id:"userCollId",val:"userName"}],canWrite:[true]},projectPerms: {ids:["proj1","proj2"],canWrite:[true, false]}, blanketPerms: 1},
    }
    return a
}

export interface SporePrintData {
    _id: string
    parent?: string // Only empty if purchased and not printed yourself
    species: string
    subspecies?: string
    creationDate: number
    color?: string
    density?: string
    pics?: PicWithNotesIncoming[]
    sale?: string
    mostRecentImage?: PicWithNotesIncoming
    notes?: Note[]
    disposed?: number
    lastUpdated: number
    acl?: ACL
}

export function SporePrintSelectorCloseable({onSelect}:{onSelect: (val?: SporePrintData)=>void}) {
    const doSel = (val?: SporePrintData):void=>{
        if (!val){
            return
        }
        onSelect(val)
    }
    return <CloseableSelector<SporePrintData> props={{
        allowCreation: false, // TODO: ok?
        doSelect: doSel, // For selecting normally
        msgTxt: ChannelTextNewAgarBatch, // TODO: ???
        closeTxt: "Close Spore Print List",
        //createTxt: "Create Fruit", // TODO: ok?
        lowercase: "sporePrint",
        //creatorInPage: sp.creatorInPage, // TODO: ok?
        //createEndpt: "fruit", // TODO: ok?
        getId: (v: FruitData) => v._id,
        createSelector:(selHdl: (onSelect: SporePrintData) => void)=>{
            return <SporePrintSelector doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: do we want a creator anywhere for this?
        // createCreator:(selHdl: (onSelect: FruitData) => void)=>{
        //     return <FruitSelector handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}