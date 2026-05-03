import {Note} from "@/app/components/formSubcomponents/notes";

export function TestGrainBatchOkFull() {
    const a: GrainBatchData = {
        _id: "(GRAIN BATCH ID HERE)",
        soakTimeHrs: 9,
        boilTimeMins: 30,
        dryTimeHours: 4,
        recipe: ("GRAIN RECIPE ID HERE"),
        creationDate: Date.now(),
        notes: [{time: Date.now(), note: "(TEST NOTE 1)"}, {time: Date.now() + 2000, note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
    }
    return a
}

export interface GrainBatchData {
    _id: string
    creationDate: number
    recipe: string
    soakTimeHrs?: number
    boilTimeMins?: number
    dryTimeHours?: number
    notes?: Note[]
    lastUpdated: number
}
