import {Note} from "@/app/components/formSubcomponents/notes";
import {ACL} from "@/app/components/accessControlServer";

export function TestLcSyringeOk(){
    let ExampleNotes;
    const a: LcSyringe = {
        _id: "(LC ID HERE)",
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        confirmedClean: undefined,
        knownFruitable: true,
        genSpore: 7,
        genFruitOrSpore: 3,
        transfersOut: ["(TRANSFER 1)","(TRANSFER 2)"],
        parent: "(PARENT ID)",
        disposed: Date.now()+40000,
        notes: ExampleNotes,
        lastUpdated: 789,
    }
    return a
}
export interface LcSyringe {
    _id: string
    parent?: string
    creationDate: number
    species: string
    subspecies?: string
    sale?: string
    genSpore?:  number
    genFruitOrSpore?: number
    confirmedClean?: boolean
    knownFruitable?: boolean
    transfersOut?: string[]
    disposed?: number
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}