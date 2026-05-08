import {Note} from "@/app/components/formSubcomponents/notes";
import {ACL} from "@/app/components/accessControlServer";

export function TestPlugsOk(){
    const testNote = ()=>{
        return {time: new Date().getTime(), note:"TEST_NOTE_TEXT_HERE"}
    }
    const testNotes: Note[] = [testNote(), testNote(), testNote()]
    const a: PlugsJar = {
        _id: "(PLATE ID HERE)",
        dowelTypes: [
            {wood:"wood1",size:1,units:"miles"},
            {wood:"wood2",size:0.5,units:"cm"},
        ],
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        innoc: "(Innoc transfer id!)",
        genSpore: 7,
        genFruitOrSpore: 3,
        parentType: "plate",
        parent: "(PARENT ID)",
        knownFruitable: true,
        sales: ["SALE_ID_HERE","SALE_ID_2_HERE"],
        disposed: Date.now()+40000,
        notes: [...testNotes],
        lastUpdated: 789,
    }
    return a
}
export interface DowelType {
    wood: string
    size: number
    units: string
}
export interface PlugsJar {
    _id: string
    parentType?: string
    parent?: string
    creationDate: number
    dowelTypes: DowelType[]
    genSpore?: number
    genFruitOrSpore?: number
    species?: string
    subspecies?: string
    innoc?: string
    transfersOut?: string[]
    knownFruitable?: boolean
    pcRun?:string
    sales?: string[]
    disposed?: number
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}