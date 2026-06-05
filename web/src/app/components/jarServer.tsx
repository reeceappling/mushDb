import {Note} from "@/app/components/formSubcomponents/notes";
import {
    ExamplePicWithNotesIncoming,
    PicWithNotesIncoming
} from "@/app/components/formSubcomponents/picWithNotes";
import {
    Contamination,
    ExampleContaminations,
    ExamplePicsWithNotesIncoming
} from "@/app/components/formSubcomponents/contaminations";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {EntryPerms} from "@/app/components/perms";
import {ACL, TestAcl} from "@/app/components/accessControlServer";
import {ChannelTextNewAgarBatch} from "@/app/components/agarBatchServer";
import {FruitingChamberSelector} from "@/app/components/fruitingChamberClient";
import {FruitingChamberData} from "@/app/components/fruitingChamberServer";
import {JarSelector} from "@/app/components/jarClient";

export function TestJarOK(){
    return new JarData({
        _id: "(JAR ID HERE)",
        sizeCups: 4,
        recipe: "(JAR RECIPE ID)",
        // grainBatch: "(GRAIN_BATCH_ID)",
        wetness: 5,
        burstGrains: 1,
        pcRun: "(PC RUN ID)",
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

export interface JarData {
    _id: string
    recipe: string // TODO: may not exist for imported jars?
    wetness?: number // TODO: handle everywhere
    burstGrains?: number // TODO: handle everywhere
    sizeCups: number
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
    contamination?: Contamination[]
    knownFruitable?: boolean
    sale?: string
    disposed?: number
    mostRecentImage?: PicWithNotesIncoming
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class JarData {
    // Accept a single object containing the fields
    constructor(init?: Partial<JarData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "jar"
    }
}

export function JarSelectorCloseable(sp: SelectorProps<JarData>) { // TODO: use
    const doSel = (val?: JarData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<JarData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Jar List",
        //createTxt: "Create Bag",// TODO: ???
        lowercase: "jar",
        //creatorInPage: sp.creatorInPage,// TODO: ???
        //createEndpt: "bag",// TODO: ???
        createSelector:(selHdl: (onSelect: JarData) => void)=>{
            return <JarSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: ok?
        // createCreator:(selHdl: (onSelect: FruitingChamberData) => void)=>{
        //     return <NewFruitingChamberForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}