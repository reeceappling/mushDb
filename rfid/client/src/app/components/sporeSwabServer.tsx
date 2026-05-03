import {Note} from "@/app/components/formSubcomponents/notes";
import {TestNotes} from "@/app/components/formSubcomponents/contaminations";
import {ACL} from "@/app/components/accessControlServer";


export function TestSporeSwabOk(){
    const a: SporeSwab = {
        _id: "(SUBSTR ID HERE)",
        parent: "(PARENT ID)",
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        sale: "SALE ID",
        disposed: Date.now(),
        notes: TestNotes,
        lastUpdated: 789,
    }
    return a
}

export interface SporeSwab {
    _id: string
    parent?: string // Only empty if purchased and not printed yourself
    parentType?:string
    creationDate: number
    species: string
    subspecies?: string
    sale?: string
    disposed?: number
    transfersOut?: string[]
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}