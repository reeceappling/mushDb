import {Note} from "@/app/components/formSubcomponents/notes";
import {
    Contamination,
} from "@/app/components/formSubcomponents/contaminations";
import {PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {ACL} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {PlateSelector} from "@/app/components/plateClient";

export interface PlateData {
    _id: string
    agarBatch?: string
    creationDate: number
    condensationCoverageAtPourTime?: number
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
    contamination?: Contamination[]
    knownFruitable?: boolean
    sale?: string
    disposed?: number
    pics?: PicWithNotesIncoming[]
    mostRecentImage?: PicWithNotesIncoming
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class PlateData {
    // Accept a single object containing the fields
    constructor(init?: Partial<PlateData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "plate"
    }
    public description(): string {
        if(this.species !== undefined){
            return `Plate ${this._id}. Species ${this.species}. ${this.subspecies!==undefined&&`Subspecies ${this.subspecies}`}. Created on ${new Date(this.creationDate).toISOString()}. Last updated on ${new Date(this.lastUpdated).toISOString()}.${this.disposed!==undefined&&` Disposed on ${new Date(this.disposed).toISOString()}`}` // TODO: KF, contams, etc?
        }
        return `Plate ${this._id}. Not innoculated. Created on ${new Date(this.creationDate).toISOString()}.${this.disposed!==undefined&&` Disposed on ${new Date(this.disposed).toISOString()}`}`
    }
}

export function PlateSelectorCloseable(sp: SelectorProps<PlateData>) { // TODO: use
    const doSel = (val?: PlateData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<PlateData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Plate List",
        lowercase: "plate",
        createSelector:(selHdl: (onSelect: PlateData) => void)=>{
            return <PlateSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
    }}/>
}