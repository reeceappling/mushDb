import {Note} from "@/app/components/formSubcomponents/notes";
import {
    ExamplePicWithNotesIncoming,
    PicWithNotesIncoming
} from "@/app/components/formSubcomponents/picWithNotes";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {EntryPerms} from "@/app/components/perms";
import {ExamplePicsWithNotesIncoming} from "@/app/components/formSubcomponents/contaminations";
import {ACL} from "@/app/components/accessControlServer";
import {ChannelTextNewAgarBatch} from "@/app/components/agarBatchServer";
import {MssSelector, NewMssForm} from "@/app/components/mssClient";
import {MssData} from "@/app/components/mssServer";
import {FruitSelector} from "@/app/components/fruitClient";

export function TestFruitOK(){
    return new FruitData({
        _id: "(FRUIT ID HERE)",
        creationDate: 1,
        species: "(SPECIES)",
        subspecies: "(SUBSPECIES)",
        genSpore: 7,
        transfersOut: ["(TRANSFER OUT 1)","(TRANSFER OUT 2)"],
        prints: ["(PRINT1)","(PRINT2)"],
        parentType: "bag",
        parent: "(PARENT ID)",
        pics: ExamplePicsWithNotesIncoming,
        disposed: Date.now()+5000,
        mostRecentImage: ExamplePicWithNotesIncoming,
        notes: [{time: Date.now(),note: "(TEST NOTE 1)"},{time: Date.now()+2000,note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
        // TODO: acl?
    })
}

export interface FruitData {
    _id: string
    creationDate: number
    species: string
    subspecies?: string
    genSpore?: number
    transfersOut?: string[]
    prints?: string[]
    parentType?: string
    parent?: string
    pics?: PicWithNotesIncoming[]
    disposed?: number
    mostRecentImage?: PicWithNotesIncoming
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}
export class FruitData {
    // Accept a single object containing the fields
    constructor(init?: Partial<FruitData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "fruit"
    }
}

export function FruitSelectorCloseable({onSelect}:{onSelect: (val?: FruitData)=>void}) {
    const doSel = (val?: FruitData):void=>{
        if (!val){
            return
        }
        onSelect(val)
    }
    return <CloseableSelector<FruitData> props={{
        allowCreation: false, // TODO: ok?
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Fruit List",
        //createTxt: "Create Fruit", // TODO: ok?
        lowercase: "fruit",
        //creatorInPage: sp.creatorInPage, // TODO: ok?
        //createEndpt: "fruit", // TODO: ok?
        createSelector:(selHdl: (onSelect: FruitData) => void)=>{
            return <FruitSelector doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: do we want a creator anywhere for this?
        // createCreator:(selHdl: (onSelect: FruitData) => void)=>{
        //     return <FruitSelector handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}