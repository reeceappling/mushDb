import {Note} from "@/app/components/formSubcomponents/notes";
import {
    Contamination,
} from "@/app/components/formSubcomponents/contaminations";
import {PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {EntryPerms} from "@/app/components/perms";
import {ACL} from "@/app/components/accessControlServer";

export function TestPlateOk(){
    const now = new Date().getTime()
    const testNote = ()=>{
        return {time: new Date().getTime(), note:"TEST_NOTE_TEXT_HERE"}
    }
    const testNotes: Note[] = [testNote(), testNote(), testNote()]
    const aPic: PicWithNotesIncoming = {time: now, notes: [...testNotes], location: "test.jpg"}
    const p: PicWithNotesIncoming[] = [aPic,aPic,aPic]
    const c: Contamination = {time: now, location: "test.jpg", mold:true, bacteria:false, confirmed:true, notes: [...testNotes]}
    const a: PlateData = {
        _id: "(PLATE ID HERE)",
        agarBatch: "(AGAR BATCH ID)", // TODO: used to be agar?
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        innoc: "(Innoc transfer id!)",
        genSpore: 7,
        genFruitOrSpore: 3,
        transfersOut: ["(TRANSFER 1)","(TRANSFER 2)"],
        parentType: "plate",
        parent: "(PARENT ID)",
        pics: [...p],
        contamination: [...[c,c]],
        knownFruitable: true,
        sale: "SALE_ID_HERE",
        disposed: Date.now()+40000,
        mostRecentImage: {...p[0]},
        notes: [...testNotes],
        lastUpdated: 789,
        //perms: {userPerms: {ids:[{id:"userCollId",val:"userName"}],canWrite:[true]},projectPerms: {ids:["proj1","proj2"],canWrite:[true, false]}, blanketPerms: 1},
    }
    return a
}
export interface PlateData {
    _id: string
    agarBatch?: string
    creationDate: number
    condensationCoverageAtSealTime?: number
    pourCoverage?: number
    wetAtCooledTime?: boolean
    agarOnOutsideAtPourTime?: boolean
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