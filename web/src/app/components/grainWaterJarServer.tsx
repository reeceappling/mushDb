import {Note} from "@/app/components/formSubcomponents/notes";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {GrainBatchSelector, NewGrainBatchForm} from "@/app/components/grainBatchClient";
import {ACL} from "@/app/components/accessControlServer";
import {GrainWaterJarSelector, NewGrainWaterJarForm} from "@/app/components/grainWaterJarClient";
import {useState} from "react";

// export function TestGrainBatchOkFull() {
//     return new GrainBatchData({
//         _id: "(GRAIN BATCH ID HERE)",
//         soakTimeHrs: 9,
//         boilTimeMins: 30,
//         dryTimeHours: 4,
//         recipe: ("GRAIN RECIPE ID HERE"),
//         creationDate: Date.now(),
//         notes: [{time: Date.now(), note: "(TEST NOTE 1)"}, {time: Date.now() + 2000, note: "(TEST NOTE 2)"}],
//         lastUpdated: 789,
//         acl: TestAcl(),
//     })
// }

export interface GrainWaterJarData {
    _id: string
    grainBatch: string
    creationDate: number
    notes?: Note[]
    lastUpdated: number
    acl: ACL
    disposed?: number
}
export class GrainWaterJarData {
    // Accept a single object containing the fields
    constructor(init?: Partial<GrainWaterJarData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "grainWaterJar" // TODO: ensure ok
    }
    public description(): string {
        return `Grain Water Jar ${this._id}. From batch ${this.grainBatch}. Created on ${new Date(this.creationDate).toISOString()}. Last updated on ${new Date(this.lastUpdated).toISOString()}`
    }
}

export function GrainWaterJarSelectorCloseable(sp: SelectorProps<GrainWaterJarData>) {
    const doSel = (val?: GrainWaterJarData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<GrainWaterJarData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Grain Water Jar List",
        createTxt: "Create Grain Water Jar",
        lowercase: "grainwater jar",
        creatorInPage: sp.creatorInPage,
        createEndpt: "grainWaterJar", // TODO: ensure ok
        createSelector:(selHdl: (onSelect: GrainWaterJarData) => void)=>{
            return <GrainWaterJarSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        createCreator:(selHdl: (onSelect: GrainWaterJarData) => void)=>{
            return <NewGrainWaterJarForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        },
    }}/>
}
export function GrainWaterJarsSelectionList({updateParent}:{updateParent:(grainWaterJars:GrainWaterJarData[],grainWaterJarsDisposed:boolean[])=>void}){
    const [currentJars, setCurrentJars] = useState<GrainWaterJarData[]>([]); // TODO: both state values may need to be in a single useState
    const [currentJarsDisposed, setCurrentJarsDisposed] = useState<boolean[]>([]);
    const addNewJar = (jar: GrainWaterJarData):void=>{
        const nextJars = [...currentJars, jar]
        const nextDisposed = [...currentJarsDisposed, false]
        setCurrentJars(nextJars)
        setCurrentJarsDisposed(nextDisposed)
        updateParent(nextJars,nextDisposed)
    }
    return <div>
        {currentJars.map((jar,index)=>{
            return <GrainWaterJarSelectionDisposeArea key={jar._id} id={jar._id} disposed={currentJarsDisposed[index]} updateParent={()=>{
                const idx = currentJars.findIndex(v=>{return v._id === jar._id})
                const currentDisp = currentJarsDisposed[idx]
                const update = [...currentJarsDisposed]
                update[idx] = !currentDisp
                setCurrentJarsDisposed(update)
                updateParent(currentJars,update)
            }} del={()=>{
                const idx = currentJars.findIndex(v=>{return v._id === jar._id})
                const nextJars = [...currentJars].filter((v,i)=>{return i!==idx})
                const nextDisposed = [...currentJarsDisposed].filter((v,i)=>{return i!==idx})
                setCurrentJars(nextJars)
                setCurrentJarsDisposed(nextDisposed)
                updateParent(nextJars,nextDisposed)
            }}/>
        })}
        <CloseableSelector<GrainWaterJarData> props={{
            allowCreation: false,
            doSelect: addNewJar, // For selecting normally
            closeTxt: "Close Grain Water Jar List",
            createTxt: "Create Grain Water Jar",
            lowercase: "grainwater jar",
            creatorInPage: false,
            createEndpt: "grainWaterJar", // TODO: ensure ok
            createSelector:(selHdl: (onSelect: GrainWaterJarData) => void)=>{
                return <GrainWaterJarSelector allowCreate={false} blacklist={currentJars.map(v=>{return v._id})} showDisposed={false} doSelect={(v)=>{
                    v && selHdl(v)
                }}/>
            },
            createCreator:(selHdl: (onSelect: GrainWaterJarData) => void)=>{
                return <NewGrainWaterJarForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
            },
        }}/>
    </div>
}
export function GrainWaterJarSelectionDisposeArea({id,disposed,updateParent,del}:{id:string,disposed:boolean,updateParent:()=>void,del:()=>void}){
    return <div>{id+"  "} {disposed?"disposed":"not disposed"}<input type={"checkbox"} onClick={updateParent}></input><button className={"removeButtonSmall"} onClick={del}>{"remove"}</button></div>
}
export function GrainWaterJarSelectorCloseableNoDisposed(sp: SelectorProps<GrainWaterJarData>) {
    const doSel = (val?: GrainWaterJarData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<GrainWaterJarData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Grain Water Jar List",
        createTxt: "Create Grain Water Jar",
        lowercase: "grainwater jar",
        creatorInPage: sp.creatorInPage,
        createEndpt: "grainWaterJar", // TODO: ensure ok
        createSelector:(selHdl: (onSelect: GrainWaterJarData) => void)=>{
            return <GrainWaterJarSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        createCreator:(selHdl: (onSelect: GrainWaterJarData) => void)=>{
            return <NewGrainWaterJarForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        },
    }}/>
}
