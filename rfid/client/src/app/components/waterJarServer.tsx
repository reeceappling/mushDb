import {Note} from "@/app/components/formSubcomponents/notes";
import {
    Contamination,
} from "@/app/components/formSubcomponents/contaminations";
import {PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {EntryPerms} from "@/app/components/perms";
import {ACL} from "@/app/components/accessControlServer";

export function TestWaterOk(){
    const now = new Date().getTime()
    const testNote = ()=>{
        return {time: new Date().getTime(), note:"TEST_NOTE_TEXT_HERE"}
    }
    const testNotes: Note[] = [testNote(), testNote(), testNote()]
    const a: WaterJarData = {
        _id: "(WATER JAR ID HERE)",
        creationDate: now,
        pcRun: "(PC RUN ID HERE)",
        notes: [...testNotes],
        lastUpdated: 789,
    }
    return a
}
export interface WaterJarData {
    _id: string
    creationDate: number
    pcRun: string
    notes?: Note[]
    disposed?: number
    lastUpdated: number
}