import {Note} from "@/app/components/formSubcomponents/notes";
import {ACL} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {NewWaterJarForm, WaterJarSelector} from "@/app/components/waterJarClient";

export interface WaterJarData {
    _id: string
    creationDate: number
    pcRun: string
    notes?: Note[]
    disposed?: number
    lastUpdated: number
    acl: ACL
}
export class WaterJarData {
    // Accept a single object containing the fields
    constructor(init?: Partial<WaterJarData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "waterJar"
    }
    public description(): string {
        const firstSent = `Water jar ${this._id}`
        const lastSent = `Created on ${new Date(this.creationDate).toISOString()}. Last updated on ${new Date(this.lastUpdated).toISOString()}.${this.disposed !== undefined && ` Disposed on ${new Date(this.disposed).toISOString()}`}`
        return `${firstSent}. ${lastSent}`
    }
}

export function WaterJarSelectorCloseable(sp: SelectorProps<WaterJarData>) {
    const doSel = (val?: WaterJarData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<WaterJarData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Water Jar List",
        createTxt: "Create Water Jar",
        lowercase: "water jar",
        creatorInPage: sp.creatorInPage,
        createEndpt: "waterJar",
        createSelector:(selHdl: (onSelect: WaterJarData) => void)=>{
            return <WaterJarSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        createCreator:(selHdl: (onSelect: WaterJarData) => void)=>{
            return <NewWaterJarForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        },
    }}/>
}