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
import {SelectorProps} from "@/app/components/selector";
import {EntryPerms} from "@/app/components/perms";
import {ACL} from "@/app/components/accessControlServer";

export function TestJarOK(){
    const a: JarData = {
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
        //perms: {userPerms: {ids:[{id:"userCollId",val:"userName"}],canWrite:[true]},projectPerms: {ids:["proj1","proj2"],canWrite:[true, false]}, blanketPerms: 1},
    }
    return a
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
    acl?: ACL
}

export function JarSelector({handlers,jar}:{handlers: SelectorProps<JarData>,jar?:JarData}){ // TODO: USE?
    // TODO: REDO! NEEDED?
    // return RecentSelector<JarData>({
    //     msgTxt: ChannelTextNewJar,
    //     recentEndpt: "jars",
    //     assertType: AssertJar,
    //     closeTxt: "Close Grain Jar List",
    //     //createTxt: "Create Fruit", // Jars only created from transfer
    //     //newForm: NewFruit, // Jars only created from transfer
    //     createEndpt: "jar",
    //     lowercase: "grain jar",
    //     inline: (inlineIn: InlineProps<JarData>)=>{return JarInline(inlineIn)},
    // })(sp)
}