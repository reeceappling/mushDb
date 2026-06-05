import {Note} from "@/app/components/formSubcomponents/notes";
import {ExamplePicWithNotesIncoming, PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {Contamination, ExampleContaminations, ExamplePicsWithNotesIncoming} from "@/app/components/formSubcomponents/contaminations";
import {EntryPerms} from "@/app/components/perms";
import {ACL, TestAcl} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {ChannelTextNewAgarBatch} from "@/app/components/agarBatchServer";
import {PlateSelector} from "@/app/components/plateClient";
import {PlateData} from "@/app/components/plateServer";
import {SlantSelector} from "@/app/components/slantClient";

export function TestSlantOk(){
    return new SlantData({
        _id: "(slant ID HERE)",
        agarBatch: "(AGAR BATCH ID)",
        stickType: "(STICK TYPE HERE)",
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        innoc: "(Innoc transfer id!)",
        genSpore: 7,
        genFruitOrSpore: 3,
        transfersOut: ["(TRANSFER 1)","(TRANSFER 2)"],
        parentType: "plate",
        parent: "(PARENT ID)",
        pics: ExamplePicsWithNotesIncoming,
        contamination: ExampleContaminations,
        knownFruitable: true,
        sale: "SALE_ID_HERE",
        disposed: Date.now()+40000,
        mostRecentImage: ExamplePicWithNotesIncoming,
        notes: [{time: Date.now(),note: "(TEST NOTE 1)"},{time: Date.now()+2000,note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
        acl: TestAcl(),
    })
}

export interface SlantData {
    _id: string
    agarBatch?: string
    stickType?: string
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
    contamination?: Contamination[]
    knownFruitable?: boolean
    sale?: string
    disposed?: number
    mostRecentImage?: PicWithNotesIncoming
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class SlantData {
    // Accept a single object containing the fields
    constructor(init?: Partial<SlantData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "slant"
    }
}

export function SlantSelectorCloseable(sp: SelectorProps<SlantData>) { // TODO: use
    const doSel = (val?: SlantData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<SlantData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Slant List",
        //createTxt: "Create Bag",// TODO: ???
        lowercase: "slant",
        //creatorInPage: sp.creatorInPage,// TODO: ???
        //createEndpt: "bag",// TODO: ???
        createSelector:(selHdl: (onSelect: SlantData) => void)=>{
            return <SlantSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: ok?
        // createCreator:(selHdl: (onSelect: FruitingChamberData) => void)=>{
        //     return <NewFruitingChamberForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}

