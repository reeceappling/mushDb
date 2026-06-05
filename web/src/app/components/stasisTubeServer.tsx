import {
    Contamination,
    ExampleContaminations,
    ExamplePicsWithNotesIncoming
} from "@/app/components/formSubcomponents/contaminations";
import {Note} from "@/app/components/formSubcomponents/notes";
import {ExamplePicWithNotesIncoming, PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {ACL, TestAcl} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {StasisTubeSelector} from "@/app/components/stasisTubeClient";

export function TestStasisTubeOk(){
    return new StasisTubeData({
        _id: "(TUBE ID HERE)",
        pcRun: "(PC RUN ID)",
        waterSource: undefined,
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
export interface StasisTubeData {
    _id: string
    pcRun?: string
    waterSource?: string
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
export class StasisTubeData {
    // Accept a single object containing the fields
    constructor(init?: Partial<StasisTubeData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "stasisTube"
    }
}

export function StasisTubneSelectorCloseable(sp: SelectorProps<StasisTubeData>) { // TODO: use
    const doSel = (val?: StasisTubeData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<StasisTubeData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Stasis Tube List",
        //createTxt: "Create Bag",// TODO: ???
        lowercase: "stasis tube",
        //creatorInPage: sp.creatorInPage,// TODO: ???
        //createEndpt: "bag",// TODO: ???
        createSelector:(selHdl: (onSelect: StasisTubeData) => void)=>{
            return <StasisTubeSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: ok?
        // createCreator:(selHdl: (onSelect: FruitingChamberData) => void)=>{
        //     return <NewFruitingChamberForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}