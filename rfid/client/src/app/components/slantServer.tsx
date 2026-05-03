import {Note} from "@/app/components/formSubcomponents/notes";
import {ExamplePicWithNotesIncoming, PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {Contamination, ExampleContaminations, ExamplePicsWithNotesIncoming} from "@/app/components/formSubcomponents/contaminations";
import {EntryPerms} from "@/app/components/perms";
import {ACL} from "@/app/components/accessControlServer";

export function TestSlantOk(){
    const a: SlantData = {
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
        //perms: {userPerms: {ids:[{id:"userCollId",val:"userName"}],canWrite:[true]},projectPerms: {ids:["proj1","proj2"],canWrite:[true, false]}, blanketPerms: 1},
    }
    return a
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
    acl?: ACL
}

// export function SlantSelector( // TODO: unlikely to need!!!!!!!!!!!!
//     {
//         doSelect, allowCreation, headerLevel, creatorInPage
//     }: {
//         doSelect:(val:SlantData | undefined)=>void
//         allowCreation?:boolean
//         headerLevel?:number
//         creatorInPage?:boolean
//     }) {
//     // TODO: THIS!
// }

