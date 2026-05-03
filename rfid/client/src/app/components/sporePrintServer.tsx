import {ExamplePicWithNotesIncoming, PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {Note} from "@/app/components/formSubcomponents/notes";
import {EntryPerms} from "@/app/components/perms";
import {ExamplePicsWithNotesIncoming, TestNotes} from "@/app/components/formSubcomponents/contaminations";
import {ACL} from "@/app/components/accessControlServer";


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